package k8s

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func WaitForPods(
	client kubernetes.Interface,
	namespace string,
	label string,
	expected int,
) error {

	for {
		pods, err := client.CoreV1().
			Pods(namespace).
			List(context.TODO(), metav1.ListOptions{
				LabelSelector: label,
			})
		if err != nil {
			return err
		}

		ready := 0
		for _, p := range pods.Items {
			if p.Status.Phase == v1.PodRunning {
				ready++
			}
		}

		if ready == expected {
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}