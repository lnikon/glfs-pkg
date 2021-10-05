package kube

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"

	// "k8s.io/apimachinery/pkg/api/errors"

	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
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

func createDeploymentClient() *v1.DeploymentInterface {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	deploymentClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)
	return &deploymentClient
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
	deploymentClient := *createDeploymentClient()
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": name,
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	fmt.Println("Kube: Creating deployment...")
	result, err := deploymentClient.Create(context.TODO(), deployment, metav1.CreateOptions{})
	if err != nil {
		log.Fatal(err)
		return err
	}

	fmt.Printf("Kube: Created deployment %q.\n", result.GetObjectMeta().GetName())
	return nil
}

func GetDeployment(name string) *appsv1.Deployment {
	deploymentClient := *createDeploymentClient()
	deployement, err := deploymentClient.Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return deployement
}

func GetAllDeployments() *appsv1.DeploymentList {
	deploymentClient := *createDeploymentClient()
	deploymentList, err := deploymentClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil
	}

	return deploymentList
}

func int32Ptr(i int32) *int32 { return &i }
