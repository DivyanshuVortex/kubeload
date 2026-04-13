package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// CreateDeployment creates a new Deployment with a single container.
// The pod template is labelled with app={name} to allow label-selector queries.
func CreateDeployment(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name, image string,
	replicas int32,
) error {
	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  name,
							Image: image,
						},
					},
				},
			},
		},
	}

	_, err := client.AppsV1().Deployments(namespace).Create(ctx, dep, metav1.CreateOptions{})
	return err
}

// DeleteDeployment deletes the named Deployment and propagates deletion to all
// owned pods before returning (ForegroundDeletion policy).
func DeleteDeployment(
	ctx context.Context,
	client kubernetes.Interface,
	namespace, name string,
) error {
	policy := metav1.DeletePropagationForeground
	return client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &policy,
	})
}
