package controllers

import (
	"context"
	"fmt"
	"reflect"
	"time"

	utils "github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/write"
	"github.com/go-logr/logr"
	appsv1alpha1 "github.com/m88i/nexus-operator/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource"
	"github.com/m88i/nexus-operator/pkg/logger"
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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/validation"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/server"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/update"
	"github.com/m88i/nexus-operator/pkg/framework"
)

var log = logger.GetLogger("controller_nexus")

const (
	updatePollWaitTimeout = 500 * time.Millisecond
	updateCancelTimeout   = 30 * time.Second
)

// NexusReconciler reconciles a Nexus object
type NexusReconciler struct {
	client.Client
	Log                logr.Logger
	Scheme             *runtime.Scheme
	DiscoveryClient    discovery.DiscoveryInterface
	ResourceSupervisor resource.Supervisor
}

// +kubebuilder:rbac:groups=apps.m88i.io,resources=nexus,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.m88i.io,resources=nexus/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core,resources=pods;services;services/finalizers;endpoints;persistentvolumeclaims;events;configmaps;secrets;replicationcontrollers;podtemplates;serviceaccounts,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=apps,resources=deployments;daemonsets;replicasets;statefulsets,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;create
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=create;delete;get;list;patch;update;watch
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=create;delete;get;list;patch;update;watch
func (r *NexusReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	ctx := context.Background()
	log.Infof("Reconciling Nexus '%s' on namespace '%s'", req.Name, req.Namespace)
	result = ctrl.Result{}

	// Fetch the Nexus instance
	instance := &appsv1alpha1.Nexus{}
	err = r.Get(ctx, req.NamespacedName, instance)
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

	v, err := validation.NewValidator(r.Client, r.Scheme, r.DiscoveryClient)
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
	err = r.ResourceSupervisor.InitManagers(validatedNexus)
	if err != nil {
		return
	}
	// Create the objects as desired by the Nexus instance
	requiredRes, err := r.ResourceSupervisor.GetRequiredResources()
	if err != nil {
		return
	}
	// Get the actual deployed objects
	deployedRes, err := r.ResourceSupervisor.GetDeployedResources()
	if err != nil {
		return
	}
	// Get the resource comparator
	comparator, err := r.ResourceSupervisor.GetComparator()
	if err != nil {
		return
	}
	deltas := comparator.Compare(deployedRes, requiredRes)

	writer := write.New(r.Client).WithOwnerController(validatedNexus, r.Scheme)
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

func (r *NexusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	b := ctrl.NewControllerManagedBy(mgr).
		For(&appsv1alpha1.Nexus{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.ServiceAccount{})

	ocp, err := openshift.IsOpenShift(r.DiscoveryClient)
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

func (r *NexusReconciler) handleUpdate(nexus *appsv1alpha1.Nexus, required, deployed map[reflect.Type][]utils.KubernetesResource) error {
	requiredDeployment := required[reflect.TypeOf(appsv1.Deployment{})][0].(*appsv1.Deployment)
	deployedDeployments := deployed[reflect.TypeOf(appsv1.Deployment{})]
	if len(deployedDeployments) == 0 {
		// nothing previously deployed, not an update
		return nil
	}
	deployedDeployment := deployedDeployments[0].(*appsv1.Deployment)
	return update.HandleUpdate(nexus, deployedDeployment, requiredDeployment, r.Scheme, r.Client)
}

func (r *NexusReconciler) ensureServerUpdates(instance *appsv1alpha1.Nexus) error {
	log.Info("Performing Nexus server operations if needed")
	status, err := server.HandleServerOperations(instance, r.Client)
	if err != nil {
		return err
	}
	log.Info("Server Operations finished. Status is %v", status)
	instance.Status.ServerOperationsStatus = status
	return nil
}

func (r *NexusReconciler) updateNexus(nexus *appsv1alpha1.Nexus, originalNexus *appsv1alpha1.Nexus, err *error) {
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
			if updateErr := r.Update(context.TODO(), nexus); errors.IsConflict(updateErr) {
				newNexus := &appsv1alpha1.Nexus{ObjectMeta: v1.ObjectMeta{
					Name:      nexus.Name,
					Namespace: nexus.Namespace,
				}}
				if err := r.Get(context.TODO(), framework.Key(newNexus), newNexus); err != nil {
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
			if updateErr := r.Status().Update(context.TODO(), nexus); errors.IsConflict(updateErr) {
				newNexus := &appsv1alpha1.Nexus{ObjectMeta: v1.ObjectMeta{
					Name:      nexus.Name,
					Namespace: nexus.Namespace,
				}}
				if err := r.Get(context.TODO(), framework.Key(newNexus), newNexus); err != nil {
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

func (r *NexusReconciler) getNexusDeploymentStatus(nexus *appsv1alpha1.Nexus) error {
	log.Info("Checking Deployment Status")
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
			log.Info("Checking Route Status")
			uri, err = openshift.GetRouteURI(r.Client, r.DiscoveryClient, types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name})
		} else if nexus.Spec.Networking.ExposeAs == appsv1alpha1.IngressExposeType {
			log.Info("Checking Ingress Status")
			uri, err = kubernetes.GetIngressURI(r.Client, types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name})
		}
		if err != nil {
			return err
		}
		nexus.Status.NexusRoute = uri
	}
	return nil
}
