package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k *Kubernetes) NamespacesList(ctx context.Context, options ResourceListOptions) (runtime.Unstructured, error) {
	return k.ResourcesList(ctx, &schema.GroupVersionKind{
		Group: "", Version: "v1", Kind: "Namespace",
	}, "", options)
}

func (k *Kubernetes) ProjectsList(ctx context.Context, options ResourceListOptions) (runtime.Unstructured, error) {
	return k.ResourcesList(ctx, &schema.GroupVersionKind{
		Group: "project.openshift.io", Version: "v1", Kind: "Project",
	}, "", options)
}

func (k *Kubernetes) NamespaceCreate(ctx context.Context, namespace string) ([]*unstructured.Unstructured, error) {
	var resources []any
	ns := &v1.Namespace{
		TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{Name: namespace, Labels: nil},
	}	
	resources = append(resources, ns)
	var toCreate []*unstructured.Unstructured
	converter := runtime.DefaultUnstructuredConverter
	for _, obj := range resources {
		m, err := converter.ToUnstructured(obj)
		if err != nil {
			return nil, err
		}
		u := &unstructured.Unstructured{}
		if err = converter.FromUnstructured(m, u); err != nil {
			return nil, err
		}
		toCreate = append(toCreate, u)
	}
	return k.resourcesCreateOrUpdate(ctx, toCreate)
}

func (k *Kubernetes) NamespaceDelete(ctx context.Context, namespace string) (string, error) {
	return namespace, k.ResourcesDelete(ctx, &schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"}, "", namespace)
}