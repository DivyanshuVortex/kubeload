package k8s

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// WaitForPods blocks until the number of ready pods matching label equals expected,
// or the context is cancelled/expired.
//
// When expected == 0, the function waits until no pods with the label exist at all
// (handles the terminating-pod window during scale-down to zero).
func WaitForPods(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, label string,
	expected int,
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pods, err := client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
			LabelSelector: label,
		})
		if err != nil {
			return err
		}

		if expected == 0 {
			// Wait until all pods (including terminating) are fully gone.
			if len(pods.Items) == 0 {
				return nil
			}
		} else {
			ready := 0
			for _, p := range pods.Items {
				if isPodReady(p) {
					ready++
				}
			}
			if ready == expected {
				return nil
			}
		}

		time.Sleep(1 * time.Second)
	}
}

// isPodReady returns true when a pod is Running, has the Ready condition True,
// and is not being terminated.
func isPodReady(pod corev1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return false
	}
	if pod.Status.Phase != corev1.PodRunning {
		return false
	}
	for _, c := range pod.Status.Conditions {
		if c.Type == corev1.PodReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}