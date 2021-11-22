package kube

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	// meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	// apiextensions "k8s.io/apiextensions-apiserver"
	"k8s.io/apimachinery/pkg/runtime/schema"
	upcxxv1alpha1types "github.com/lnikon/glfs-pkg/pkg/upcxx-operator/api/v1alpha1"
	glconst "github.com/lnikon/glfs-pkg/pkg/constants"
	upcxxv1alpha1clientset "github.com/lnikon/glfs-pkg/pkg/upcxx-operator/clientset/v1alpha1"
	// uuid "k8s.io/apimachinery/pkg/util/uuid"
)

func init() {
	if flag.Lookup("kubeconfig") == nil {
		if home := homedir.HomeDir(); home != "" {
			flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
		} else {
			flag.String("kubeconfig", "", "")
		}
	}
}

func createUpcxxClient() upcxxv1alpha1clientset.UPCXXInterface {
	flag.Parse()
	kubeconfig := flag.Lookup("kubeconfig").Value.(flag.Getter).Get().(string)
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil || config == nil {
		log.Fatal(err)
	}

	clientset, err := upcxxv1alpha1clientset.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	return clientset.UPCXX("default")
}

func GetPodsCount() int {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Fatal(err.Error())
	}

	return len(pods.Items)
}

func CreateUPCXX(name string) error {
	upcxxClient := createUpcxxClient()

	groupVersionKind := schema.GroupVersionKind{}
	groupVersionKind.Group = upcxxv1alpha1types.GroupVersion.Group
	groupVersionKind.Version = upcxxv1alpha1types.GroupVersion.Version
	groupVersionKind.Kind = "UPCXX"

	apiVersion, kind := groupVersionKind.ToAPIVersionAndKind()

	log.Default().Printf("%v\n", groupVersionKind)

	upcxx := &upcxxv1alpha1types.UPCXX{
		TypeMeta: metav1.TypeMeta{
			Kind: kind,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta {
			Name: name,
			Namespace: "default",
			// OwnerReferences: []meta.OwnerReference{
			// 	{
			// 		APIVersion: apiVersion,
			// 		Kind: kind,
			//     UID: uuid.NewUUID(),
			// 		Name: name,
			// 	},
			// },
		},
		Spec: upcxxv1alpha1types.UPCXXSpec{
			StatefulSetName: name,
			WorkerCount: 2,
			Algorithm: glconst.Kruskal,
		},
		Status: upcxxv1alpha1types.UPCXXStatus{},
	}

	upcxx, err := upcxxClient.Create(upcxx)
	return err
}

func GetDeployment(name string) *upcxxv1alpha1types.UPCXX {
	upcxxClient := createUpcxxClient()
	deployement, err := upcxxClient.Get(name, metav1.GetOptions{})
	if err != nil {
		return nil
	}

	return deployement
}

func GetAllDeployments() *upcxxv1alpha1types.UPCXXList {
	deploymentClient := createUpcxxClient()
	deploymentList, err := deploymentClient.List(metav1.ListOptions{})
	if err != nil {
		return nil
	}

	return deploymentList
}

