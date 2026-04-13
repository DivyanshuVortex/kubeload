package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ScaleDeployment(
	client kubernetes.Interface,
	namespace string,
	name string,
	replicas int32,
) error {

	dep, err := client.AppsV1().
		Deployments(namespace).
		Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	dep.Spec.Replicas = &replicas

	_, err = client.AppsV1().
		Deployments(namespace).
		Update(context.TODO(), dep, metav1.UpdateOptions{})

	return err
}