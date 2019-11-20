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
	"github.com/m88i/nexus-operator/pkg/openshift"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"reflect"

	utilres "github.com/RHsyseng/operator-utils/pkg/resource"
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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_nexus")

var resourceMgr resource.NexusResourceManager

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
	reconcileNexus.resourceManager = resource.New(reconcileNexus.client, reconcileNexus.discoveryClient)
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

	watchOwnedObjects := []runtime.Object{
		&appsv1.Deployment{},
		&corev1.Service{},
		&corev1.PersistentVolumeClaim{},
		&routev1.Route{},
	}
	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.Nexus{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			if isNoKindMatchError(routev1.GroupVersion.Group, err) {
				// ignore if Route API is not found
				continue
			}
			return err
		}
	}
	return nil
}

// isNoKindMatchError verify if the given error is NoKindMatchError for the given group
func isNoKindMatchError(group string, err error) bool {
	if kindErr, ok := err.(*meta.NoKindMatchError); ok {
		return kindErr.GroupKind.Group == group
	}
	return false
}

// blank assignment to verify that ReconcileNexus implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNexus{}

// ReconcileNexus reconciles a Nexus object
type ReconcileNexus struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client          client.Client
	scheme          *runtime.Scheme
	discoveryClient discovery.DiscoveryInterface
	resourceManager resource.NexusResourceManager
}

// Reconcile reads that state of the cluster for a Nexus object and makes changes based on the state read
// and what is in the Nexus.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNexus) Reconcile(request reconcile.Request) (result reconcile.Result, err error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Nexus")
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

	deployedRes := make(map[reflect.Type][]utilres.KubernetesResource)

	// In case of any errors from here, we should update the application status
	defer r.updateNexusStatus(instance, &result, &err)

	// Create the objects as desired by the Nexus instance
	requestedRes, err := r.resourceManager.CreateRequiredResources(instance)
	if err != nil {
		return
	}
	// Get the actual deployed objects
	deployedRes, err = r.resourceManager.GetDeployedResources(instance)
	if err != nil {
		return
	}

	comparator := resource.GetComparator()
	deltas := comparator.Compare(deployedRes, requestedRes)

	writer := write.New(r.client).WithOwnerController(instance, r.scheme)
	for resourceType, delta := range deltas {
		if !delta.HasChanges() {
			continue
		}
		log.Info("Will ",
			"create", len(delta.Added),
			"update", len(delta.Updated),
			"delete", len(delta.Removed),
			"instances of", resourceType)
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

func (r *ReconcileNexus) updateNexusStatus(nexus *appsv1alpha1.Nexus, result *reconcile.Result, err *error) {
	log.Info("Updating application status before leaving")

	if *err != nil {
		nexus.Status.NexusStatus = fmt.Sprintf("Failed to deploy Nexus: %s", *err)
	} else {
		nexus.Status.NexusStatus = "OK"
	}

	if statusErr := r.updateNexusDeploymentStatus(nexus); statusErr != nil {
		log.Error(statusErr, "Error while fetching Nexus Deployment status")
		err = &statusErr
	}

	if routeErr := r.updateNexusRouteStatus(nexus); routeErr != nil {
		log.Error(routeErr, "Error while fetching Nexus Route status")
		err = &routeErr
	}

	if updateErr := r.client.Status().Update(context.TODO(), nexus); updateErr != nil {
		log.Error(updateErr, "Error while updating Nexus status")
		err = &updateErr
	}

	log.Info("Controller finished reconciliation")
}

func (r *ReconcileNexus) updateNexusDeploymentStatus(nexus *appsv1alpha1.Nexus) error {
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

func (r *ReconcileNexus) updateNexusRouteStatus(nexus *appsv1alpha1.Nexus) error {
	log.Info("Checking Route Status")
	uri, err := openshift.GetRouteURI(r.client, r.discoveryClient, types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name})
	if err != nil {
		return err
	}
	nexus.Status.NexusRoute = uri
	return nil
}
