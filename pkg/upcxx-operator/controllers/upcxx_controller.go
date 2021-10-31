/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	pgasv1alpha1 "github.com/lnikon/upcxx-operator/api/v1alpha1"
)

const (
	// Container that contains UPCXX graphs library and application
	UPCXXContainerName       = "upcxx"
	UPCXXContainerTagLatest  = ":latest"
	UPCXXLatestContainerName = UPCXXContainerName + UPCXXContainerTagLatest
)

// UPCXXReconciler reconciles a UPCXX object
type UPCXXReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger
}

//+kubebuilder:rbac:groups=pgas.github.com,resources=upcxxes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pgas.github.com,resources=upcxxes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pgas.github.com,resources=upcxxes/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the UPCXX object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *UPCXXReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("UPCXX", req.NamespacedName)

	// 1. Check if my resource kind exists
	log.Info("Fetching UPCXX resource")
	upcxx := pgasv1alpha1.UPCXX{}
	if err := r.Client.Get(ctx, req.NamespacedName, &upcxx); err != nil {
		log.Error(err, "Resource already exsits")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 2. Get UPCXXJob with given name
	log = log.WithValues("statefulset_name", upcxx.Spec.StatefulSetName)
	deployment := apps.Deployment{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: upcxx.Namespace, Name: upcxx.Spec.StatefulSetName}, &deployment)
	if apierrors.IsNotFound(err) {
		log.Info("Could not find existing deployment for", "resource", upcxx.Spec.StatefulSetName)

		deployment = *buildDeployment(upcxx)
		if err := r.Client.Create(ctx, &deployment); err != nil {
			log.Error(err, "Failed to create Deployment", "resource", upcxx.Spec.StatefulSetName)
			return ctrl.Result{}, nil
		}

		r.Recorder.Eventf(&upcxx, core.EventTypeNormal, "Created", "deployment", &deployment.Name)
		log.Info("Created Deployment resource for UPCXX")
		return ctrl.Result{}, nil
	}

	if err != nil {
		log.Error(err, "Failed to get Deployment resource for UPCXX")
		return ctrl.Result{}, err
	}

	// 3. If exists, ignore and apply changed
	log.Info("Deployment already exists for UPCXX", "resource", upcxx.Spec.StatefulSetName)

	expectedReplicas := upcxx.Spec.WorkerCount
	if expectedReplicas != *deployment.Spec.Replicas {
		log.Info("Updating replica count for Deployment of", "resource", upcxx.Spec.StatefulSetName)
		deployment.Spec.Replicas = &expectedReplicas
		if err := r.Client.Update(ctx, &deployment); err != nil {
			log.Error(err, "Failed to update", "old_replica_count", deployment.Spec.Replicas, "new_replica_count", expectedReplicas, "resource", upcxx.Spec.StatefulSetName)
			return ctrl.Result{}, err
		}

		log.Info("Successfuly updated replica count for", "resource", upcxx.Spec.StatefulSetName)
	}

	return ctrl.Result{}, nil
}

func buildDeployment(upcxx pgasv1alpha1.UPCXX) *apps.Deployment {
	deployment := apps.Deployment{
		ObjectMeta: meta.ObjectMeta{
			Name:            upcxx.Spec.StatefulSetName,
			Namespace:       upcxx.Namespace,
			OwnerReferences: []meta.OwnerReference{*meta.NewControllerRef(&upcxx, pgasv1alpha1.GroupVersion.WithKind("UPCXX"))},
		},
		Spec: apps.DeploymentSpec{
			Replicas: &upcxx.Spec.WorkerCount,
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"example-controller.jetstack.io/deployment-name": upcxx.Spec.StatefulSetName,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Labels: map[string]string{
						"example-controller.jetstack.io/deployment-name": upcxx.Spec.StatefulSetName,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
	return &deployment
}

// SetupWithManager sets up the controller with the Manager.
func (r *UPCXXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pgasv1alpha1.UPCXX{}).
		Complete(r)
}
