package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"kubeload/internal/config"
	"kubeload/internal/k8s"
	"kubeload/internal/load"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var cfg config.Config

var rootCmd = &cobra.Command{
	Use:   "kubeload",
	Short: "Kubernetes scale latency load generator",
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the load generator",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithTimeout(cmd.Context(), cfg.Timeout)
		defer cancel()

		// Intercept Ctrl+C and SIGTERM: cancel context so workers exit
		// cleanly and deferred cleanup can run before the process exits.
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			select {
			case <-sigCh:
				fmt.Println("\nInterrupted — cancelling and cleaning up...")
				cancel()
			case <-ctx.Done():
			}
		}()

		// ── Connect ─────────────────────────────────────────────────────────────
		client, err := k8s.NewClient()
		if err != nil {
			fmt.Println("error creating client:", err)
			os.Exit(1)
		}

		_, err = client.CoreV1().Pods(cfg.Namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			fmt.Println("cluster unreachable:", err)
			os.Exit(1)
		}
		fmt.Println("Connected to cluster")

		// ── Create one deployment per worker ────────────────────────────────────
		deployments := make([]string, cfg.Workers)
		for i := 0; i < cfg.Workers; i++ {
			name := fmt.Sprintf("%s-%d", cfg.DeploymentName, i)
			deployments[i] = name
			fmt.Printf("Creating deployment %s (image: %s, min replicas: %d)\n",
				name, cfg.Image, cfg.MinReplicas)
			if err := k8s.CreateDeployment(ctx, client, cfg.Namespace, name, cfg.Image, int32(cfg.MinReplicas)); err != nil {
				fmt.Printf("failed to create deployment %s: %v\n", name, err)
				os.Exit(1)
			}
		}

		// ── Cleanup ─────────────────────────────────────────────────────────────
		if cfg.Cleanup {
			defer func() {
				fmt.Println("\nCleaning up deployments...")
				// Use a fresh context: the run context may already be cancelled.
				delCtx, delCancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer delCancel()
				for _, name := range deployments {
					if delErr := k8s.DeleteDeployment(delCtx, client, cfg.Namespace, name); delErr != nil {
						fmt.Printf("  warning: failed to delete %s: %v\n", name, delErr)
					} else {
						fmt.Printf("  deleted %s\n", name)
					}
				}
			}()
		}

		// ── Run workers concurrently ─────────────────────────────────────────────
		fmt.Printf("\nRunning %d worker(s) × %d cycle(s) (min=%d → max=%d replicas, dwell=%s)\n\n",
			cfg.Workers, cfg.Cycles, cfg.MinReplicas, cfg.MaxReplicas, cfg.Dwell)

		var mu sync.Mutex
		var allSamples []load.Sample
		var wg sync.WaitGroup

		for i := 0; i < cfg.Workers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				wcfg := load.WorkerConfig{
					WorkerID:       workerID,
					DeploymentName: deployments[workerID],
					Namespace:      cfg.Namespace,
					MinReplicas:    cfg.MinReplicas,
					MaxReplicas:    cfg.MaxReplicas,
					Cycles:         cfg.Cycles,
					Dwell:          cfg.Dwell,
				}
				samples := load.RunWorker(ctx, client, wcfg)
				mu.Lock()
				allSamples = append(allSamples, samples...)
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// ── Print metrics ────────────────────────────────────────────────────────
		fmt.Printf("\n=== Results: %d worker(s) × %d cycle(s) ===",
			cfg.Workers, cfg.Cycles)
		load.Report(allSamples)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	runCmd.Flags().StringVar(&cfg.DeploymentName, "name", "kubeload", "Base name for deployments (each worker appends -N)")
	runCmd.Flags().StringVar(&cfg.Namespace, "namespace", "default", "Kubernetes namespace")
	runCmd.Flags().StringVar(&cfg.Image, "image", "nginx:alpine", "Container image for test pods")
	runCmd.Flags().IntVar(&cfg.MinReplicas, "min-replicas", 0, "Replica count to scale down to")
	runCmd.Flags().IntVar(&cfg.MaxReplicas, "max-replicas", 3, "Replica count to scale up to")
	runCmd.Flags().IntVar(&cfg.Workers, "workers", 1, "Concurrent workers (each gets its own deployment)")
	runCmd.Flags().IntVar(&cfg.Cycles, "cycles", 3, "Scale up→down cycles per worker")
	runCmd.Flags().DurationVar(&cfg.Dwell, "dwell", 2*time.Second, "Hold time at max-replicas before scaling down")
	runCmd.Flags().DurationVar(&cfg.Timeout, "timeout", 10*time.Minute, "Global run timeout")
	runCmd.Flags().BoolVar(&cfg.Cleanup, "cleanup", true, "Delete deployments after run")

	rootCmd.AddCommand(runCmd)
}
