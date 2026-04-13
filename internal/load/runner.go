package load

import (
	"context"
	"fmt"
	"time"

	"kubeload/internal/k8s"

	"k8s.io/client-go/kubernetes"
)

// WorkerConfig holds parameters for a single concurrent worker.
type WorkerConfig struct {
	WorkerID       int
	DeploymentName string
	Namespace      string
	MinReplicas    int
	MaxReplicas    int
	Cycles         int
	Dwell          time.Duration
}

// RunWorker executes Cycles iterations of scale-up → dwell → scale-down for one worker.
// It returns all collected samples regardless of errors.
func RunWorker(ctx context.Context, client kubernetes.Interface, cfg WorkerConfig) []Sample {
	label := fmt.Sprintf("app=%s", cfg.DeploymentName)
	samples := make([]Sample, 0, cfg.Cycles*2)

	for cycle := 1; cycle <= cfg.Cycles; cycle++ {
		if ctx.Err() != nil {
			break
		}

		// ── Scale up ────────────────────────────────────────────────────────────
		t0 := time.Now()
		err := k8s.ScaleDeployment(ctx, client, cfg.Namespace, cfg.DeploymentName, int32(cfg.MaxReplicas))
		if err == nil {
			err = k8s.WaitForPods(ctx, client, cfg.Namespace, label, cfg.MaxReplicas)
		}
		upLatency := time.Since(t0)
		samples = append(samples, Sample{
			Worker:    cfg.WorkerID,
			Cycle:     cycle,
			Direction: "up",
			From:      cfg.MinReplicas,
			To:        cfg.MaxReplicas,
			Latency:   upLatency,
			Err:       err,
		})
		if err != nil {
			fmt.Printf("[worker %d] cycle %d: scale-up error: %v\n", cfg.WorkerID, cycle, err)
			continue
		}
		fmt.Printf("[worker %d] cycle %d: up   → %d pods ready in %s\n",
			cfg.WorkerID, cycle, cfg.MaxReplicas, fmtDur(upLatency))

		// ── Dwell ───────────────────────────────────────────────────────────────
		select {
		case <-ctx.Done():
			return samples
		case <-time.After(cfg.Dwell):
		}

		// ── Scale down ──────────────────────────────────────────────────────────
		t0 = time.Now()
		err = k8s.ScaleDeployment(ctx, client, cfg.Namespace, cfg.DeploymentName, int32(cfg.MinReplicas))
		if err == nil {
			err = k8s.WaitForPods(ctx, client, cfg.Namespace, label, cfg.MinReplicas)
		}
		downLatency := time.Since(t0)
		samples = append(samples, Sample{
			Worker:    cfg.WorkerID,
			Cycle:     cycle,
			Direction: "down",
			From:      cfg.MaxReplicas,
			To:        cfg.MinReplicas,
			Latency:   downLatency,
			Err:       err,
		})
		if err != nil {
			fmt.Printf("[worker %d] cycle %d: scale-down error: %v\n", cfg.WorkerID, cycle, err)
			continue
		}
		fmt.Printf("[worker %d] cycle %d: down → %d pods ready in %s\n",
			cfg.WorkerID, cycle, cfg.MinReplicas, fmtDur(downLatency))
	}

	return samples
}