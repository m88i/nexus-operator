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
	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	routev1 "github.com/openshift/api/route/v1"
	networking "k8s.io/api/networking/v1beta1"
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
		&networking.Ingress{},
	}

	ownerHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1alpha1.Nexus{},
	}
	for _, watchObject := range watchOwnedObjects {
		err = c.Watch(&source.Kind{Type: watchObject}, ownerHandler)
		if err != nil {
			if isNoKindMatchError(routev1.GroupVersion.Group, err) ||
				isNoKindMatchError(networking.GroupName, err) {
				// ignore if Route or Ingress API is not found
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

	// default networking parameters
	if err = r.setDefaultNetworking(instance); err != nil {
		return
	}

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

// setDefaultNetworking verify given CR parameters for networking and set defaults
func (r *ReconcileNexus) setDefaultNetworking(nexus *appsv1alpha1.Nexus) (err error) {
	if !nexus.Spec.Networking.Expose {
		return nil
	}

	ocp := false
	if ocp, err = openshift.IsOpenShift(r.discoveryClient); err != nil {
		return err
	}

	if ocp {
		if nexus.Spec.Networking.ExposeAs == appsv1alpha1.IngressExposeType {
			return fmt.Errorf("Ingress is only available on Kubernetes. Try '%s' as the expose type ", appsv1alpha1.RouteExposeType)
		}
		if len(nexus.Spec.Networking.ExposeAs) == 0 {
			nexus.Spec.Networking.ExposeAs = appsv1alpha1.RouteExposeType
		}
	} else {
		if nexus.Spec.Networking.ExposeAs == appsv1alpha1.RouteExposeType {
			return fmt.Errorf("Routes is only available on OpenShift. Try '%s' as the expose type ", appsv1alpha1.IngressExposeType)
		}
		if len(nexus.Spec.Networking.ExposeAs) == 0 {
			nexus.Spec.Networking.ExposeAs = appsv1alpha1.IngressExposeType
		}
	}

	if nexus.Spec.Networking.ExposeAs == appsv1alpha1.NodePortExposeType && nexus.Spec.Networking.NodePort == 0 {
		return fmt.Errorf("NodePort networking requires a port. Check Nexus resource 'spec.networking.nodePort' parameter ")
	}

	if nexus.Spec.Networking.ExposeAs == appsv1alpha1.IngressExposeType && len(nexus.Spec.Networking.Host) == 0 {
		return fmt.Errorf("Ingress networking requires a host. Check Nexus resource 'spec.networking.host' parameter ")
	}

	return nil
}

func (r *ReconcileNexus) updateNexusStatus(nexus *appsv1alpha1.Nexus, result *reconcile.Result, err *error) {
	log.Info("Updating application status before leaving")
	cache := nexus.DeepCopy()

	if *err != nil {
		nexus.Status.NexusStatus = fmt.Sprintf("Failed to deploy Nexus: %s", *err)
	} else {
		nexus.Status.NexusStatus = okStatus
	}

	if statusErr := r.getNexusDeploymentStatus(nexus); statusErr != nil {
		log.Error(statusErr, "Error while fetching Nexus Deployment status")
		err = &statusErr
	}

	if urlErr := r.getNexusURL(nexus); urlErr != nil {
		log.Error(urlErr, "Error while fetching Nexus URL status")
		err = &urlErr
	}

	if !reflect.DeepEqual(cache, nexus) {
		log.Info("Updating nexus status")
		if updateErr := r.client.Update(context.TODO(), nexus); updateErr != nil {
			log.Error(updateErr, "Error while updating Nexus status")
			err = &updateErr
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
