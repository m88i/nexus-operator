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

package nexus

import (
	"context"
	"fmt"
	"reflect"
	"time"

	resUtils "github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1alpha1 "github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/validation"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/server"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/update"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/m88i/nexus-operator/pkg/logger"
)

var log = logger.GetLogger("controller_nexus")

const (
	updatePollWaitTimeout = 500 * time.Millisecond
	updateCancelTimeout   = 30 * time.Second
)

// Add creates a new Nexus Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	reconcileNexus := &ReconcileNexus{
		client:          mgr.GetClient(),
		discoveryClient: discovery.NewDiscoveryClientForConfigOrDie(mgr.GetConfig()),
		scheme:          mgr.GetScheme(),
	}
	reconcileNexus.resourceSupervisor = resource.NewSupervisor(reconcileNexus.client, reconcileNexus.discoveryClient)
	return reconcileNexus
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("nexus-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Nexus
	err = c.Watch(&source.Kind{Type: &appsv1alpha1.Nexus{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	controllerWatcher := framework.NewControllerWatcher(r.(*ReconcileNexus).discoveryClient, mgr, c, &appsv1alpha1.Nexus{})
	watchedObjects := []framework.WatchedObjects{
		{
			GroupVersion: routev1.GroupVersion,
			AddToScheme:  routev1.Install,
			Objects:      []runtime.Object{&routev1.Route{}},
		},
		{
			GroupVersion: networking.SchemeGroupVersion,
			AddToScheme:  networking.AddToScheme,
			Objects:      []runtime.Object{&networking.Ingress{}},
		},
		{Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}, &corev1.PersistentVolumeClaim{}, &corev1.ServiceAccount{}}},
	}
	if err = controllerWatcher.Watch(watchedObjects...); err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileNexus implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNexus{}

// ReconcileNexus reconciles a Nexus object
type ReconcileNexus struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client             client.Client
	scheme             *runtime.Scheme
	discoveryClient    discovery.DiscoveryInterface
	resourceSupervisor resource.Supervisor
}

// Reconcile reads that state of the cluster for a Nexus object and makes changes based on the state read
// and what is in the Nexus.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNexus) Reconcile(request reconcile.Request) (result reconcile.Result, err error) {
	log.Infof("Reconciling Nexus '%s' on namespace '%s'", request.Name, request.Namespace)
	result = reconcile.Result{}

	// Fetch the Nexus instance
	instance := &appsv1alpha1.Nexus{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
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

	v, err := validation.NewValidator(r.client, r.scheme, r.discoveryClient)
	if err != nil {
		// Error using the discovery API - requeue the request.
		return
	}

	validatedNexus, err := v.SetDefaultsAndValidate(instance)
	// In case of any errors from here, we should update the Nexus CR and its status
	defer r.updateNexus(validatedNexus, instance, &err)
	if err != nil {
		return
	}

	// Initialize the resource managers
	err = r.resourceSupervisor.InitManagers(validatedNexus)
	if err != nil {
		return
	}
	// Create the objects as desired by the Nexus instance
	requiredRes, err := r.resourceSupervisor.GetRequiredResources()
	if err != nil {
		return
	}
	// Get the actual deployed objects
	deployedRes, err := r.resourceSupervisor.GetDeployedResources()
	if err != nil {
		return
	}
	// Get the resource comparator
	comparator, err := r.resourceSupervisor.GetComparator()
	if err != nil {
		return
	}
	deltas := comparator.Compare(deployedRes, requiredRes)

	writer := write.New(r.client).WithOwnerController(validatedNexus, r.scheme)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Info("Will ",
			"create ", len(delta.Added),
			", update ", len(delta.Updated),
			", delete ", len(delta.Removed),
			" instances of ", resourceType)
		_, err = writer.AddResources(delta.Added)
		if err != nil {
			return
		}
		_, err = writer.UpdateResources(deployedRes[resourceType], delta.Updated)
		if err != nil {
			return
		}
		_, err = writer.RemoveResources(delta.Removed)
		if err != nil {
			return
		}
	}

	if err = r.ensureServerUpdates(validatedNexus); err != nil {
		return
	}

	// Check if we are performing an update and act upon it if needed
	err = r.handleUpdate(validatedNexus, requiredRes, deployedRes)
	return
}

func (r *ReconcileNexus) handleUpdate(nexus *appsv1alpha1.Nexus, required, deployed map[reflect.Type][]resUtils.KubernetesResource) error {
	requiredDeployment := required[reflect.TypeOf(appsv1.Deployment{})][0].(*appsv1.Deployment)
	deployedDeployments := deployed[reflect.TypeOf(appsv1.Deployment{})]
	if len(deployedDeployments) == 0 {
		// nothing previously deployed, not an update
		return nil
	}
	deployedDeployment := deployedDeployments[0].(*appsv1.Deployment)
	return update.HandleUpdate(nexus, deployedDeployment, requiredDeployment, r.scheme, r.client)
}

func (r *ReconcileNexus) ensureServerUpdates(instance *appsv1alpha1.Nexus) error {
	log.Info("Performing Nexus server operations if needed")
	status, err := server.HandleServerOperations(instance, r.client)
	if err != nil {
		return err
	}
	log.Infof("Server Operations finished. Status is %v", status)
	instance.Status.ServerOperationsStatus = status
	return nil
}

func (r *ReconcileNexus) updateNexus(nexus *appsv1alpha1.Nexus, originalNexus *appsv1alpha1.Nexus, err *error) {
	log.Info("Updating application status before leaving")

	if statusErr := r.getNexusDeploymentStatus(nexus); statusErr != nil {
		log.Errorf("Error while fetching Nexus Deployment status: %v", err)
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
		log.Errorf("Error while fetching Nexus URL status: %v", urlErr)
	}

	if !reflect.DeepEqual(originalNexus.Spec, nexus.Spec) {
		log.Infof("Updating Nexus instance '%s'", nexus.Name)
		waitErr := wait.Poll(updatePollWaitTimeout, updateCancelTimeout, func() (bool, error) {
			if updateErr := r.client.Update(context.TODO(), nexus); errors.IsConflict(updateErr) {
				newNexus := &appsv1alpha1.Nexus{ObjectMeta: v1.ObjectMeta{
					Name:      nexus.Name,
					Namespace: nexus.Namespace,
				}}
				if err := r.client.Get(context.TODO(), framework.Key(newNexus), newNexus); err != nil {
					return false, err
				}
				// we override only the spec, which we are interested into
				newNexus.Spec = nexus.Spec
				nexus = newNexus
				return false, nil
			} else if updateErr != nil {
				return false, updateErr
			}
			return true, nil
		})
		if waitErr != nil {
			log.Error(waitErr, "Error while updating Nexus status")
		}
	}

	if !reflect.DeepEqual(originalNexus.Status, nexus.Status) {
		log.Infof("Updating status for Nexus instance '%s'", nexus.Name)
		waitErr := wait.Poll(updatePollWaitTimeout, updateCancelTimeout, func() (bool, error) {
			if updateErr := r.client.Status().Update(context.TODO(), nexus); errors.IsConflict(updateErr) {
				newNexus := &appsv1alpha1.Nexus{ObjectMeta: v1.ObjectMeta{
					Name:      nexus.Name,
					Namespace: nexus.Namespace,
				}}
				if err := r.client.Get(context.TODO(), framework.Key(newNexus), newNexus); err != nil {
					return false, err
				}
				// we override only the spec, which we are interested into
				newNexus.Status = nexus.Status
				nexus = newNexus
				return false, nil
			} else if updateErr != nil {
				return false, updateErr
			}
			return true, nil
		})
		if waitErr != nil {
			log.Error(waitErr, "Error while updating Nexus status")
		}
	}

	log.Info("Controller finished reconciliation")
}

func (r *ReconcileNexus) getNexusDeploymentStatus(nexus *appsv1alpha1.Nexus) error {
	log.Info("Checking Deployment Status")
	deployment := &appsv1.Deployment{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name}, deployment); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
	}
	nexus.Status.DeploymentStatus = deployment.Status
	return nil
}

func (r *ReconcileNexus) getNexusURL(nexus *appsv1alpha1.Nexus) error {
	if nexus.Spec.Networking.Expose {
		var err error
		uri := ""
		if nexus.Spec.Networking.ExposeAs == appsv1alpha1.RouteExposeType {
			log.Info("Checking Route Status")
			uri, err = openshift.GetRouteURI(r.client, r.discoveryClient, types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name})
		} else if nexus.Spec.Networking.ExposeAs == appsv1alpha1.IngressExposeType {
			log.Info("Checking Ingress Status")
			uri, err = kubernetes.GetIngressURI(r.client, types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name})
		}
		if err != nil {
			return err
		}
		nexus.Status.NexusRoute = uri
	}
	return nil
}
