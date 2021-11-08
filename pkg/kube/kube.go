package kube

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	// "k8s.io/apimachinery/pkg/api/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	upcxxv1alpha1types "github.com/lnikon/glfs-pkg/pkg/upcxx-operator/api/v1alpha1"
	upcxxv1alpha1clientset "github.com/lnikon/glfs-pkg/pkg/upcxx-operator/clientset/v1alpha1"
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

	return clientset.UPCXX("")
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

func CreateDeployment(name string) error {
	// upcxx := upcxxv1alpha1.UPCXX{}
	// upcxx.Spec.StatefulSetName = name
	// upcxx.Spec.WorkerCount = 2

	// deploymentClient := *createDeploymentClient()
	// deployment := &appsv1.Deployment{
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name: name,
	// 	},
	// 	Spec: appsv1.DeploymentSpec{
	// 		Replicas: int32Ptr(2),
	// 		Selector: &metav1.LabelSelector{
	// 			MatchLabels: map[string]string{
	// 				"app": name,
	// 			},
	// 		},
	// 		Template: apiv1.PodTemplateSpec{
	// 			ObjectMeta: metav1.ObjectMeta{
	// 				Labels: map[string]string{
	// 					"app": name,
	// 				},
	// 			},
	// 			Spec: apiv1.PodSpec{
	// 				Containers: []apiv1.Container{
	// 					{
	// 						Name:  "web",
	// 						Image: "nginx:1.12",
	// 						Ports: []apiv1.ContainerPort{
	// 							{
	// 								Name:          "http",
	// 								Protocol:      apiv1.ProtocolTCP,
	// 								ContainerPort: 80,
	// 							},
	// 						},
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// }

	// fmt.Println("Kube: Creating deployment...")
	// result, err := deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	// if err != nil {
	// 	log.Fatal(err)
	// 	return err
	// }

	// fmt.Printf("Kube: Created deployment %q.\n", result.GetObjectMeta().GetName())
	// return nil
	return nil
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

// func int32Ptr(i int32) *int32 { return &i }
