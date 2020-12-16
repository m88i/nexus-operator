// Copyright 2020 Nexus Operator and/or its authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	resUtils "github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	"github.com/go-logr/logr"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1alpha1 "github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource"
	"github.com/m88i/nexus-operator/controllers/nexus/server"
	"github.com/m88i/nexus-operator/controllers/nexus/update"
	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	"github.com/m88i/nexus-operator/pkg/framework"
)

const (
	updatePollWaitTimeout = 500 * time.Millisecond
	updateCancelTimeout   = 30 * time.Second
)

// NexusReconciler reconciles a Nexus object
type NexusReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	Supervisor resource.Supervisor
}

// +kubebuilder:rbac:groups=apps.m88i.io,resources=nexus,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.m88i.io,resources=nexus/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.m88i.io,resources=nexus/finalizers,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=services;persistentvolumeclaims;events;secrets;serviceaccounts,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;create
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get
// +kubebuilder:rbac:groups=apps,resources=deployments/finalizers,verbs=update
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=create;delete;get;list;patch;update;watch

func (r *NexusReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("nexus", req.NamespacedName)

	r.Log.Info("Reconciling Nexus")
	result := ctrl.Result{}

	// Fetch the Nexus instance
	nexus := &appsv1alpha1.Nexus{}
	err := r.Get(context.TODO(), req.NamespacedName, nexus)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return result, nil
		}
		// Error reading the object - requeue the request.
		return result, err
	}

	// In case of any errors from here, we should update the Nexus CR and its status
	defer r.updateNexus(nexus, &err)

	// Initialize the resource managers
	err = r.Supervisor.InitManagers(nexus)
	if err != nil {
		return result, err
	}
	// Create the objects as desired by the Nexus instance
	requiredRes, err := r.Supervisor.GetRequiredResources()
	if err != nil {
		return result, err
	}
	// Get the actual deployed objects
	deployedRes, err := r.Supervisor.GetDeployedResources()
	if err != nil {
		return result, err
	}
	// Get the resource comparator
	comparator, err := r.Supervisor.GetComparator()
	if err != nil {
		return result, err
	}
	deltas := comparator.Compare(deployedRes, requiredRes)

	writer := write.New(r).WithOwnerController(nexus, r.Scheme)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		r.Log.Info("Will ",
			"create ", len(delta.Added),
			", update ", len(delta.Updated),
			", delete ", len(delta.Removed),
			" instances of ", resourceType)
		_, err = writer.AddResources(delta.Added)
		if err != nil {
			return result, err
		}
		_, err = writer.UpdateResources(deployedRes[resourceType], delta.Updated)
		if err != nil {
			return result, err
		}
		_, err = writer.RemoveResources(delta.Removed)
		if err != nil {
			return result, err
		}
	}

	if err = r.ensureServerUpdates(nexus); err != nil {
		return result, err
	}

	// Check if we are performing an update and act upon it if needed
	err = r.handleUpdate(nexus, requiredRes, deployedRes)
	return result, err
}

func (r *NexusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Nexus{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ServiceAccount{})

	ocp, err := discovery.IsOpenShift()
	if err != nil {
		return err
	}
	if ocp {
		b.Owns(&routev1.Route{})
	} else {
		b.Owns(&networking.Ingress{})
	}
	return b.Complete(r)
}

func (r *NexusReconciler) handleUpdate(nexus *appsv1alpha1.Nexus, required, deployed map[reflect.Type][]resUtils.KubernetesResource) error {
	requiredDeployment := required[reflect.TypeOf(appsv1.Deployment{})][0].(*appsv1.Deployment)
	deployedDeployments := deployed[reflect.TypeOf(appsv1.Deployment{})]
	if len(deployedDeployments) == 0 {
		// nothing previously deployed, not an update
		return nil
	}
	deployedDeployment := deployedDeployments[0].(*appsv1.Deployment)
	return update.HandleUpdate(nexus, deployedDeployment, requiredDeployment, r.Scheme, r)
}

func (r *NexusReconciler) ensureServerUpdates(instance *appsv1alpha1.Nexus) error {
	r.Log.Info("Performing Nexus server operations if needed")
	status, err := server.HandleServerOperations(instance, r)
	if err != nil {
		return err
	}
	r.Log.Info("Server Operations finished", "Status", status)
	instance.Status.ServerOperationsStatus = status
	return nil
}

func (r *NexusReconciler) updateNexus(nexus *appsv1alpha1.Nexus, err *error) {
	r.Log.Info("Updating application status before leaving")
	originalNexus := nexus.DeepCopy()

	if statusErr := r.getNexusDeploymentStatus(nexus); statusErr != nil {
		r.Log.Error(*err, "Error while fetching Nexus Deployment status")
	}

	if *err != nil {
		nexus.Status.Reason = fmt.Sprintf("Failed to deploy Nexus: %s", *err)
		nexus.Status.NexusStatus = appsv1alpha1.NexusStatusFailure
	} else {
		nexus.Status.Reason = ""
		if nexus.Status.DeploymentStatus.AvailableReplicas == nexus.Spec.Replicas {
			nexus.Status.NexusStatus = appsv1alpha1.NexusStatusOK
		} else {
			nexus.Status.NexusStatus = appsv1alpha1.NexusStatusPending
		}
	}

	if urlErr := r.getNexusURL(nexus); urlErr != nil {
		r.Log.Error(urlErr, "Error while fetching Nexus URL status")
	}

	if !reflect.DeepEqual(originalNexus.Status, nexus.Status) {
		r.Log.Info("Updating status for ", "Nexus instance", nexus.Name)
		waitErr := wait.Poll(updatePollWaitTimeout, updateCancelTimeout, func() (bool, error) {
			if updateErr := r.Status().Update(context.TODO(), nexus); errors.IsConflict(updateErr) {
				newNexus := &appsv1alpha1.Nexus{ObjectMeta: v1.ObjectMeta{
					Name:      nexus.Name,
					Namespace: nexus.Namespace,
				}}
				if err := r.Get(context.TODO(), framework.Key(newNexus), newNexus); err != nil {
					return false, err
				}
				// we override only the status, which we are interested into
				newNexus.Status = nexus.Status
				nexus = newNexus
				return false, nil
			} else if updateErr != nil {
				return false, updateErr
			}
			return true, nil
		})
		if waitErr != nil {
			r.Log.Error(waitErr, "Error while updating Nexus status")
		}
	}

	r.Log.Info("Controller finished reconciliation")
}

func (r *NexusReconciler) getNexusDeploymentStatus(nexus *appsv1alpha1.Nexus) error {
	r.Log.Info("Checking Deployment Status")
	deployment := &appsv1.Deployment{}
	if err := r.Get(context.TODO(), types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name}, deployment); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}
	nexus.Status.DeploymentStatus = deployment.Status
	return nil
}

func (r *NexusReconciler) getNexusURL(nexus *appsv1alpha1.Nexus) error {
	if nexus.Spec.Networking.Expose {
		var err error
		uri := ""
		if nexus.Spec.Networking.ExposeAs == appsv1alpha1.RouteExposeType {
			r.Log.Info("Checking Route Status")
			uri, err = openshift.GetRouteURI(r, types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name})
		} else if nexus.Spec.Networking.ExposeAs == appsv1alpha1.IngressExposeType {
			r.Log.Info("Checking Ingress Status")
			uri, err = kubernetes.GetIngressURI(r, types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name})
		}
		if err != nil {
			return err
		}
		nexus.Status.NexusRoute = uri
	}
	return nil
}
