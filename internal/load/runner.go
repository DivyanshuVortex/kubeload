package load

import (
	"fmt"
	"time"

	"kubeload/internal/k8s"

	"k8s.io/client-go/kubernetes"
)

func Run(
	client kubernetes.Interface,
	namespace string,
	deployment string,
	label string,
	maxReplicas int,
	delay time.Duration,
) error {

	for i := 1; i <= maxReplicas; i++ {

		fmt.Println("Scaling to", i)

		err := k8s.ScaleDeployment(
			client,
			namespace,
			deployment,
			int32(i),
		)
		if err != nil {
			return err
		}

		err = k8s.WaitForPods(
			client,
			namespace,
			label,
			i,
		)
		if err != nil {
			return err
		}

		fmt.Println("Pods ready:", i)

		time.Sleep(delay)
	}

	return nil
}