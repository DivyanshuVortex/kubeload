package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

// ScaleDeployment updates the replica count of an existing Deployment.
// It retries automatically on conflict (optimistic concurrency collision),
// re-fetching the latest resourceVersion before each attempt.
func ScaleDeployment(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name string,
	replicas int32,
) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		dep, err := client.AppsV1().
			Deployments(namespace).
			Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		dep.Spec.Replicas = &replicas

		_, err = client.AppsV1().
			Deployments(namespace).
			Update(ctx, dep, metav1.UpdateOptions{})

		return err
	})
}