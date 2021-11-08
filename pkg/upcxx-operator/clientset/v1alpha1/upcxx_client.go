package v1alpha1

import (
	"context"

	"github.com/lnikon/glfs-pkg/pkg/upcxx-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *UPCXXClient) List(opts metav1.ListOptions) (*v1alpha1.UPCXXList, error) {
	result := v1alpha1.UPCXXList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("upcxxes").
		VersionedParams(&opts, metav1.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}

func (c *UPCXXClient) Get(name string, opts metav1.GetOptions) (*v1alpha1.UPCXX, error) {
	result := v1alpha1.UPCXX{}
	err := c.restClient.
		Get().
		Namespace("default").
		Resource("upcxxes").
		Name(name).
		VersionedParams(&opts, metav1.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}
