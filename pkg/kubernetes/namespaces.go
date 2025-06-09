package kubernetes

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k *Kubernetes) NamespacesList(ctx context.Context) (string, error) {
	return k.ResourcesList(ctx, &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Namespace",
	}, "")
}

func (k *Kubernetes) NamespaceGet(ctx context.Context, namespace string) (string, error) {
	return k.ResourcesGet(ctx, &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Namespace",
	}, "", namespace)
}

func (k *Kubernetes) ProjectsList(ctx context.Context) (string, error) {
	return k.ResourcesList(ctx, &schema.GroupVersionKind{
		Group: "project.openshift.io", Version: "v1", Kind: "Project",
	}, "")
}

func (k *Kubernetes) ProjectGet(ctx context.Context, project string) (string, error) {
	return k.ResourcesGet(ctx, &schema.GroupVersionKind{
		Group: "project.openshift.io", Version: "v1", Kind: "Project",
	}, "", project)
}

func (k *Kubernetes) NamespaceCreate(ctx context.Context, namespace string, opts metav1.CreateOptions) (*corev1.Namespace, error) {
	return k.clientSet.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespace}}, opts)
}

func (k *Kubernetes) NamespaceDelete(ctx context.Context, namespace string, opts metav1.DeleteOptions) error {
	return k.clientSet.CoreV1().Namespaces().Delete(ctx, namespace, opts)
}
