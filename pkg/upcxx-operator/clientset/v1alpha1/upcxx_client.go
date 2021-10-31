package v1alpha1

import (
	"context"

	"github.com/lnikon/glfs-pkg/pkg/upcxx-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type UPCXXInterface interface {
	List(opts metav1.ListOptions) (*v1alpha1.UPCXXList, error)
}

type UPCXXClient struct {
	restClient rest.Interface
	ns         string
}

func (c *UPCXXClient) List(opts metav1.ListOptions) (*v1alpha1.UPCXXList, error) {
	result := v1alpha1.UPCXXList{}
	err := c.restClient.
		Get().
		Namespace(c.ns).
		Resource("upcxx").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do(context.TODO()).
		Into(&result)

	return &result, err
}
