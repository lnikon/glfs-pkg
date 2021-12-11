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

	"github.com/go-logr/logr"
	apps "k8s.io/api/apps/v1"
	batch "k8s.io/api/batch/v1"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	pgasv1alpha1 "github.com/lnikon/glfs-pkg/pkg/upcxx-operator/api/v1alpha1"

	"fmt"
	"strings"
)

const (
	// Container that contains UPCXX graphs library and application
	UPCXXContainerName       = "pgasgraph"
	UPCXXContainerTagLatest  = ":v0.1"
	UPCXXLatestContainerName = UPCXXContainerName + UPCXXContainerTagLatest

	// Launcher specific definitions
	launcherSuffix        = "-launcher"
	launcherJobSuffix     = "-launcher-job"
	launcherServiceSuffix = "-launcher-service"

	// Worker specific definitions
	workerSuffix        = "-worker"
	workerServiceSuffix = "-worker-service"
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

	upcxx := pgasv1alpha1.UPCXX{}
	if err := r.Client.Get(ctx, req.NamespacedName, &upcxx); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
	}

	launcherService := &core.Service{}
	err := r.Client.Get(ctx, client.ObjectKey{Namespace: upcxx.Namespace, Name: buildLauncherJobName(&upcxx)}, launcherService)
	if apierrors.IsNotFound(err) {
		log.Info("Could not find existing Service for launcher Job")

		launcherService = buildLauncherService(&upcxx)
		if err := r.Client.Create(ctx, launcherService); err != nil {
			log.Error(err, "Unable to create Service for Launcher Job")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&upcxx, core.EventTypeNormal, "Created Service for launcher Job", buildLauncherJobName(&upcxx))
	}

	workerService := &core.Service{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: upcxx.Namespace, Name: buildWorkerPodName(&upcxx)}, workerService)
	if apierrors.IsNotFound(err) {
		log.Info("Could not find existing Service for launcher Job")

		workerService = buildWorkerService(&upcxx)
		if err := r.Client.Create(ctx, workerService); err != nil {
			log.Error(err, "Unable to create Service for Launcher Job")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&upcxx, core.EventTypeNormal, "Created Service for worker StatefulSet", buildWorkerPodName(&upcxx))
	}

	log = log.WithValues("StatefulSetName", upcxx.Spec.StatefulSetName)
	statefulSet := &apps.StatefulSet{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: upcxx.Namespace, Name: buildWorkerPodName(&upcxx)}, statefulSet)
	if apierrors.IsNotFound(err) {
		log.Info("Could not find existing StatefulSet for", "resource", buildWorkerPodName(&upcxx))
		statefulSet = buildWorkerStatefulSet(&upcxx)

		if err := r.Client.Create(ctx, statefulSet); err != nil {
			log.Error(err, "Failed to create StatefulSet", "resource", buildWorkerPodName(&upcxx))
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&upcxx, core.EventTypeNormal, "Created StatefulSet", buildWorkerPodName(&upcxx))
	}

	launcherJob := &batch.Job{}
	err = r.Client.Get(ctx, client.ObjectKey{Namespace: upcxx.Namespace, Name: buildLauncherJobName(&upcxx)}, launcherJob)
	if apierrors.IsNotFound(err) {
		log.Info("Could not find existing Job for launcher job")

		launcherJob = buildLauncherJob(&upcxx)
		if err := r.Client.Create(ctx, launcherJob); err != nil {
			log.Error(err, "Failed to create Job for launcher pod")
			return ctrl.Result{}, err
		}

		r.Recorder.Eventf(&upcxx, core.EventTypeNormal, "Created Job for launcher", buildLauncherJobName(&upcxx))
	}

	// podList := &core.PodList{}
	// err = r.List(ctx, podList, client.InNamespace(upcxx.Namespace), client.MatchingLabels{"ownerStatefulSet": upcxx.Spec.StatefulSetName})
	// if err != nil {
	// 	log.Error(err, "unable to get pod list for UPCXX resource")
	// }

	// expectedReplicas := upcxx.Spec.WorkerCount
	// if expectedReplicas != *statefulSet.Spec.Replicas {
	// 	log.Info("Updating replica count forDeployment of", "resource", upcxx.Spec.StatefulSetName)
	// 	statefulSet.Spec.Replicas = &expectedReplicas
	// 	if err := r.Client.Update(ctx, statefulSet); err != nil {
	// 		log.Error(err, "Failed to update", "old_replica_count", statefulSet.Spec.Replicas, "new_replica_count", expectedReplicas, "resource", upcxx.Spec.StatefulSetName)
	// 		return ctrl.Result{}, err
	// 	}

	// 	log.Info("Successfuly updated replica count for", "resource", upcxx.Spec.StatefulSetName)
	// }

	return ctrl.Result{}, nil
}

func buildLauncherJobName(upcxx *pgasv1alpha1.UPCXX) string {
	return upcxx.Spec.StatefulSetName + launcherJobSuffix
}

func buildLauncherPodName(upcxx *pgasv1alpha1.UPCXX) string {
	return upcxx.Spec.StatefulSetName + launcherSuffix
}

func buildLauncherJob(upcxx *pgasv1alpha1.UPCXX) *batch.Job {
	controllerRef := *meta.NewControllerRef(upcxx, pgasv1alpha1.GroupVersion.WithKind("UPCXX"))
	launcherJobSpec := &batch.Job{
		ObjectMeta: meta.ObjectMeta{
			Name:      buildLauncherJobName(upcxx),
			Namespace: upcxx.ObjectMeta.Namespace,
			Labels: map[string]string{
				"app": buildLauncherJobName(upcxx),
			},
			OwnerReferences: []meta.OwnerReference{controllerRef},
		},
		Spec: batch.JobSpec{
			// TTLSecondsAfterFinished: mpiJob.Spec.RunPolicy.TTLSecondsAfterFinished,
			// ActiveDeadlineSeconds:   mpiJob.Spec.RunPolicy.ActiveDeadlineSeconds,
			BackoffLimit: int32ToPtr(1),
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Name: buildLauncherJobName(upcxx),
					Labels: map[string]string{
						"app": buildLauncherJobName(upcxx),
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            UPCXXContainerName,
							Image:           UPCXXLatestContainerName,
							ImagePullPolicy: "Never",
							Command:         []string{"sleep", "infinity"},
							Ports: []core.ContainerPort{
								{
									Name:          buildLauncherJobName(upcxx),
									ContainerPort: 80,
								},
							},
						},
					},
					RestartPolicy: core.RestartPolicyNever,
				},
			},
		},
	}

	launcherJobSpec.Spec.Template.Spec.Containers[0].Env = append(launcherJobSpec.Spec.Template.Spec.Containers[0].Env, createSSHServersEnv(upcxx.Spec.StatefulSetName, upcxx.Spec.WorkerCount))
	launcherJobSpec.Spec.Template.Spec.Containers[0].Env = append(launcherJobSpec.Spec.Template.Spec.Containers[0].Env, core.EnvVar{Name: "UPCXX_NETWORK", Value: "udp"})

	return launcherJobSpec
}

func buildWorkerPodName(upcxx *pgasv1alpha1.UPCXX) string {
	return upcxx.Spec.StatefulSetName + workerSuffix
}

func buildWorkerStatefulSet(upcxx *pgasv1alpha1.UPCXX) *apps.StatefulSet {
	controllerRef := *meta.NewControllerRef(upcxx, pgasv1alpha1.GroupVersion.WithKind("UPCXX"))
	statefulSet := apps.StatefulSet{
		ObjectMeta: meta.ObjectMeta{
			// TODO: Should we pass sts name in the yaml? It can be same as the resource name or with -sts postfix.
			Name:            buildWorkerPodName(upcxx),
			Namespace:       upcxx.Namespace,
			OwnerReferences: []meta.OwnerReference{controllerRef},
		},
		Spec: apps.StatefulSetSpec{
			ServiceName: buildWorkerPodName(upcxx),
			Replicas:    getWorkerCount(upcxx),
			Selector: &meta.LabelSelector{
				MatchLabels: map[string]string{
					"app": buildWorkerPodName(upcxx),
				},
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Name: buildWorkerPodName(upcxx),
					Labels: map[string]string{
						"app": buildWorkerPodName(upcxx),
					},
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:            UPCXXContainerName,
							Image:           UPCXXLatestContainerName,
							ImagePullPolicy: "Never",
							Command:         []string{"sleep", "infinity"},
							Ports: []core.ContainerPort{
								{
									Name:          buildWorkerPodName(upcxx),
									ContainerPort: 80,
								},
							},
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

	statefulSet.Spec.Template.Spec.Containers[0].Env = append(statefulSet.Spec.Template.Spec.Containers[0].Env, createSSHServersEnv(upcxx.Spec.StatefulSetName, upcxx.Spec.WorkerCount))
	statefulSet.Spec.Template.Spec.Containers[0].Env = append(statefulSet.Spec.Template.Spec.Containers[0].Env, core.EnvVar{Name: "UPCXX_NETWORK", Value: "udp"})

	return &statefulSet
}

func createSSHServersEnv(stsName string, replicas int32) core.EnvVar {
	sshServersList := []string{}
	for idx := int32(0); idx < replicas; idx++ {
		sshServersList = append(sshServersList, fmt.Sprintf("%s-%d", stsName, idx))
	}

	return core.EnvVar{Name: "SSH_SERVERS", Value: strings.Join(sshServersList, ",")}
}

func getWorkerCount(upcxx *pgasv1alpha1.UPCXX) *int32 {
	workerCount := upcxx.Spec.WorkerCount - 1
	return &workerCount
}

func buildLauncherService(upcxx *pgasv1alpha1.UPCXX) *core.Service {
	return newService(upcxx, buildLauncherJobName(upcxx))
}

func buildWorkerService(upcxx *pgasv1alpha1.UPCXX) *core.Service {
	return newService(upcxx, buildWorkerPodName(upcxx))
}

func newService(upcxx *pgasv1alpha1.UPCXX, name string) *core.Service {
	return &core.Service{
		ObjectMeta: meta.ObjectMeta{
			Name:      name,
			Namespace: upcxx.Namespace,
			Labels: map[string]string{
				"app": name,
			},
			OwnerReferences: []meta.OwnerReference{
				*meta.NewControllerRef(upcxx, pgasv1alpha1.GroupVersion.WithKind("UPCXX")),
			},
		},
		Spec: core.ServiceSpec{
			Ports: []core.ServicePort{
				{
					Name: name,
					Port: 80,
				},
			},
			ClusterIP: core.ClusterIPNone,
			Selector: map[string]string{
				"app": name,
			},
		},
	}
}

func int32ToPtr(i int32) *int32 {
	return &i
}

// SetupWithManager sets up the controller with the Manager.
func (r *UPCXXReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pgasv1alpha1.UPCXX{}).
		Complete(r)
}
