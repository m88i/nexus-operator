//     Copyright 2019 Nexus Operator and/or its authors
//
//     This file is part of Nexus Operator.
//
//     Nexus Operator is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     Nexus Operator is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with Nexus Operator.  If not, see <https://www.gnu.org/licenses/>.

package nexus

import (
	"context"
	"fmt"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/validation"
	"reflect"

	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/m88i/nexus-operator/pkg/logger"
	routev1 "github.com/openshift/api/route/v1"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"

	"github.com/RHsyseng/operator-utils/pkg/resource/write"

	appsv1alpha1 "github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logger.GetLogger("controller_nexus")

var watchedObjects = []framework.WatchedObjects{
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

const okStatus = "OK"

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

	v, err := validation.NewValidator(r.discoveryClient)
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
	requestedRes, err := r.resourceSupervisor.GetRequiredResources()
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
	deltas := comparator.Compare(deployedRes, requestedRes)

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

	return
}

func (r *ReconcileNexus) updateNexus(nexus *appsv1alpha1.Nexus, originalNexus *appsv1alpha1.Nexus, err *error) {
	log.Info("Updating application status before leaving")

	if *err != nil {
		nexus.Status.NexusStatus = fmt.Sprintf("Failed to deploy Nexus: %s", *err)
	} else {
		nexus.Status.NexusStatus = okStatus
	}

	if statusErr := r.getNexusDeploymentStatus(nexus); statusErr != nil {
		log.Error(statusErr, "Error while fetching Nexus Deployment status")
	}

	if urlErr := r.getNexusURL(nexus); urlErr != nil {
		log.Error(urlErr, "Error while fetching Nexus URL status")
	}

	if !reflect.DeepEqual(originalNexus, nexus) {
		log.Info("Updating nexus status")
		if updateErr := r.client.Update(context.TODO(), nexus); updateErr != nil {
			log.Error(updateErr, "Error while updating Nexus status")
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
