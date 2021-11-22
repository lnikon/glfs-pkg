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
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	pgasv1alpha1 "github.com/lnikon/glfs-pkg/pkg/upcxx-operator/api/v1alpha1"
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

	// 1. Check if UPCXX resource exists
	upcxx := pgasv1alpha1.UPCXX{}
	if err := r.Client.Get(ctx, req.NamespacedName, &upcxx); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	if upcxx.OwnerReferences == nil {
		log.Info("Owner reference for UPCXX is nil!")
		upcxx.OwnerReferences = []meta.OwnerReference{*meta.NewControllerRef(&upcxx, pgasv1alpha1.GroupVersion.WithKind("UPCXX"))}
	}

	log.Info("[VAGAG]", "upcxx.OwnerReferences", &upcxx.OwnerReferences)

	// log.Info("[VAGAG] UPCXX", "upcxx", upcxx)

	// 2. Get UPCXXJob with given name
	log = log.WithValues("StatefulSetName", upcxx.Spec.StatefulSetName)
	statefulSet := &apps.StatefulSet{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: upcxx.Namespace, Name: upcxx.Spec.StatefulSetName}, statefulSet)
	if apierrors.IsNotFound(err) {
		log.Info("Could not find existing StatefulSet for", "resource", upcxx.Spec.StatefulSetName)
		statefulSet = buildStatefulSet(upcxx)

		log.Info("[VAGAG 1] StatefulSet", "StatefulSet", &statefulSet)

		// Maybe this client creates empty one? Use StatefulSet client from the clientset to check!
		if err := r.Client.Create(ctx, statefulSet); err != nil {
			log.Error(err, "Failed to create StatefulSet", "resource", upcxx.Spec.StatefulSetName)
			return ctrl.Result{}, nil
		}

		log.Info("[VAGAG 2] StatefulSet", "StatefulSet", &statefulSet)

		r.Recorder.Eventf(&upcxx, core.EventTypeNormal, "Created StatefulSet", statefulSet.Name)
		log.Info("Created StatefulSet resource for UPCXX")
	} else {
		// log.Info("StatefulSet already exists for UPCXX", "StatefulSet", statefulSet)
	}

	// if err != nil {
	// 	log.Error(err, "Failed to get StatefulSet resource for UPCXX")
	// 	return ctrl.Result{}, err
	// }

	// 3. If exists, ignore and apply changes
	expectedReplicas := upcxx.Spec.WorkerCount
	if expectedReplicas != *statefulSet.Spec.Replicas {
		log.Info("Updating replica count for Deployment of", "resource", upcxx.Spec.StatefulSetName)
		statefulSet.Spec.Replicas = &expectedReplicas
		if err := r.Client.Update(ctx, statefulSet); err != nil {
			log.Error(err, "Failed to update", "old_replica_count", statefulSet.Spec.Replicas, "new_replica_count", expectedReplicas, "resource", upcxx.Spec.StatefulSetName)
			return ctrl.Result{}, err
		}

		log.Info("Successfuly updated replica count for", "resource", upcxx.Spec.StatefulSetName)
	}

	return ctrl.Result{}, nil
}

func buildStatefulSet(upcxx pgasv1alpha1.UPCXX) *apps.StatefulSet {
	controllerRef := *meta.NewControllerRef(&upcxx, pgasv1alpha1.GroupVersion.WithKind("UPCXX"))
	statefulSet := apps.StatefulSet{
		TypeMeta: meta.TypeMeta{},
		ObjectMeta: meta.ObjectMeta{
			Name:            upcxx.Spec.StatefulSetName,
			Namespace:       upcxx.Namespace,
			OwnerReferences: []meta.OwnerReference{controllerRef},
		},
		Spec: apps.StatefulSetSpec{
			Replicas: &upcxx.Spec.WorkerCount,
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app": upcxx.Spec.StatefulSetName,
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Name: upcxx.Spec.StatefulSetName,
					Labels: map[string]string{
						"app": upcxx.Spec.StatefulSetName,
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            "nginx",
							Image:           "nginx",
							ImagePullPolicy: "Always",
							VolumeMounts: []core.VolumeMount{
								{
									Name:      upcxx.Spec.StatefulSetName + "-vm",
									MountPath: "/vmount",
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []core.PersistentVolumeClaim{
				{
					ObjectMeta: meta.ObjectMeta{
						Name:            upcxx.Spec.StatefulSetName + "-vm",
						Namespace:       upcxx.Namespace,
						OwnerReferences: []meta.OwnerReference{controllerRef},
					},
					Spec: core.PersistentVolumeClaimSpec{
						AccessModes: []core.PersistentVolumeAccessMode{
							core.ReadWriteOnce,
						},
						Resources: core.ResourceRequirements{
							Requests: core.ResourceList{
								core.ResourceStorage: *resource.NewQuantity(500, resource.BinarySI),
							},
						},
					},
				},
			},
		},
	}

	return &statefulSet
}

// SetupWithManager sets up the controller with the Manager.
func (r *UPCXXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pgasv1alpha1.UPCXX{}).
		Complete(r)
}
