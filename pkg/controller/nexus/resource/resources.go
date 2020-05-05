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

package resource

import (
	"context"
	"fmt"
	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/discovery"

	"github.com/RHsyseng/operator-utils/pkg/resource"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
)

// NexusResourceManager is the resources manager for the nexus CR.
// Handles the creation of every single resource needed to deploy a nexus server instance on Kubernetes
type NexusResourceManager interface {
	// GetDeployedResources will fetch for the resources managed by the nexus instance deployed in the cluster
	GetDeployedResources(nexus *v1alpha1.Nexus) (resources map[reflect.Type][]resource.KubernetesResource, err error)
	// CreateRequiredResources will create the requests resources as it's supposed to be
	CreateRequiredResources(nexus *v1alpha1.Nexus) (resources map[reflect.Type][]resource.KubernetesResource, err error)
}

type resourceA interface {
	runtime.Object
	v12.ObjectMetaAccessor
}

type nexusResourceManager struct {
	NexusResourceManager
	client          client.Client
	discoveryClient discovery.DiscoveryInterface
}

// New creates a new resource manager for nexus CR
func New(client client.Client, discoveryClient discovery.DiscoveryInterface) NexusResourceManager {
	return &nexusResourceManager{
		client:          client,
		discoveryClient: discoveryClient,
	}
}

func (r *nexusResourceManager) GetDeployedResources(nexus *v1alpha1.Nexus) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	reader := read.New(r.client).WithNamespace(nexus.Namespace).WithOwnerObject(nexus)

	if routeAvailable, routeErr := openshift.IsRouteAvailable(r.discoveryClient); routeErr != nil {
		return nil, routeErr
	} else if routeAvailable {
		resources, err = reader.ListAll(&v1.PersistentVolumeClaimList{}, &v1.ServiceList{}, &appsv1.DeploymentList{}, &routev1.RouteList{})
	} else {
		resources, err = reader.ListAll(&v1.PersistentVolumeClaimList{}, &v1.ServiceList{}, &appsv1.DeploymentList{})
	}

	// Necessary to support < 1.14 K8s clusters
	if available, err := kubernetes.IsIngressAvailable(r.discoveryClient); err != nil {
		return nil, err
	} else if available {
		// ingress cannot be listed using utils, so we load the single one
		ingressResType := reflect.TypeOf(v1beta1.Ingress{})
		ingress, err := reader.Load(ingressResType, nexus.Name)
		if err == nil {
			resources[ingressResType] = []resource.KubernetesResource{ingress}
		}
	}

	if err != nil && !errors.IsNotFound(err) {
		log.Error(err, "Failed to fetch deployed nexus resources")
		return nil, err
	}
	if resources == nil {
		log.Info("No deployed resources found")
	}
	log.Infof("Number of deployed resources %d", len(resources))
	return resources, nil
}

func (r *nexusResourceManager) CreateRequiredResources(nexus *v1alpha1.Nexus) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	log.Infof("Creating resources structures namespace: %s, name: %s", nexus.Namespace, nexus.Name)
	var pvc *v1.PersistentVolumeClaim
	resources = make(map[reflect.Type][]resource.KubernetesResource)
	service := newService(nexus)

	if nexus.Spec.Persistence.Persistent {
		pvc = newPVC(nexus)
		resources[reflect.TypeOf(v1.PersistentVolumeClaim{})] = []resource.KubernetesResource{pvc}
	}

	if nexus.Spec.Networking.Expose {
		switch nexus.Spec.Networking.ExposeAs {
		case v1alpha1.RouteExposeType:
			route, err := r.createRoute(nexus, service)
			if err != nil {
				return nil, err
			}
			resources[reflect.TypeOf(routev1.Route{})] = []resource.KubernetesResource{route}
		case v1alpha1.IngressExposeType:
			ingress, err := r.createIngress(nexus, service)
			if err != nil {
				return nil, err
			}
			resources[reflect.TypeOf(v1beta1.Ingress{})] = []resource.KubernetesResource{ingress}
		}
	}

	resources[reflect.TypeOf(appsv1.Deployment{})] = []resource.KubernetesResource{newDeployment(nexus, pvc)}
	resources[reflect.TypeOf(v1.Service{})] = []resource.KubernetesResource{service}
	return resources, nil
}

func (r *nexusResourceManager) createIngress(nexus *v1alpha1.Nexus, service *v1.Service) (*v1beta1.Ingress, error) {
	// Necessary to support < 1.14 K8s clusters
	if available, err := kubernetes.IsIngressAvailable(r.discoveryClient); err != nil {
		return nil, err
	} else if !available {
		return nil, fmt.Errorf("ingress is not available in this cluster")
	}
	builder := (&ingressBuilder{}).newIngress(nexus, service)

	if len(nexus.Spec.Networking.TLS.SecretName) > 0 {
		if err := r.updateWatchedSecret(nexus); err != nil {
			return nil, err
		}
		builder = builder.withCustomTLS()
	}

	ingress, err := builder.build()
	if err != nil {
		return nil, fmt.Errorf("couldn't create ingress: %v", err)
	}
	return ingress, nil
}

func (r *nexusResourceManager) createRoute(nexus *v1alpha1.Nexus, service *v1.Service) (*routev1.Route, error) {
	if available, err := openshift.IsRouteAvailable(r.discoveryClient); err != nil {
		return nil, err
	} else if !available {
		return nil, fmt.Errorf("route not available")
	}

	builder := (&routeBuilder{}).newRoute(nexus, service)

	if nexus.Spec.Networking.TLS.Mandatory {
		builder = builder.withRedirect()
	}

	route, err := builder.build()
	if err != nil {
		return nil, fmt.Errorf("couldn't create route: %v", err)
	}
	return route, nil
}

func (r *nexusResourceManager) updateWatchedSecret(nexus *v1alpha1.Nexus) error {
	secret := &v1.Secret{}
	objectKey := types.NamespacedName{Name: nexus.Spec.Networking.TLS.SecretName, Namespace: nexus.Namespace}

	err := r.client.Get(context.TODO(), objectKey, secret)
	if err != nil {
		return fmt.Errorf("couldn't fetch Secret (%s): %v", objectKey.Name, err)
	}

	oldIngress := &v1beta1.Ingress{}
	objectKey = types.NamespacedName{Namespace: nexus.Namespace, Name: nexus.Name}

	err = r.client.Get(context.TODO(), objectKey, oldIngress)
	if err != nil {
		// there is no previous Ingress deployed, so we only need to take ownership of the secret
		if errors.IsNotFound(err) {
			err = r.ownNewObject(nexus, nil, secret)
			if err != nil {
				return fmt.Errorf("couldn't take ownership of new secret: %v", err)
			}
		} else {
			return fmt.Errorf("couldn't fetch deployed Ingress (%s): %v", objectKey.Name, err)
		}
	}
	// Previously deployed Ingress has no TLS configuration, so we only need to take ownership of the secret
	if len(oldIngress.Spec.TLS) == 0 {
		err = r.ownNewObject(nexus, nil, secret)
		if err != nil {
			return fmt.Errorf("couldn't take ownership of new secret: %v", err)
		}
	}
	// Previously deployed Ingress already has TLS configuration, but it's not using the same secret
	// We need to remove ownership of previous secret and add to new one
	if len(oldIngress.Spec.TLS) == 1 && oldIngress.Spec.TLS[0].SecretName != nexus.Spec.Networking.TLS.SecretName {
		oldSecret := &v1.Secret{}
		objectKey := types.NamespacedName{Name: oldIngress.Spec.TLS[0].SecretName, Namespace: nexus.Namespace}

		err := r.client.Get(context.TODO(), objectKey, oldSecret)
		if err != nil {
			if !errors.IsNotFound(err) {
				return fmt.Errorf("coulnd't fetch previous secret: %v", err)
			} else {
				// The previous secret has been deleted, so we don't need to adjust it
				err = r.ownNewObject(nexus, nil, secret)
				if err != nil {
					return fmt.Errorf("couldn't take ownership of new secret: %v", err)
				}
			}
		}
		err = r.ownNewObject(nexus, oldSecret, secret)
		if err != nil {
			return fmt.Errorf("couldn't take ownership of new secret: %v", err)
		}
	}
	return nil
}

func (r *nexusResourceManager) ownNewObject(nexus *v1alpha1.Nexus, oldObject, newObject resource.KubernetesResource) error {
	if oldObject != nil {
		err := removeOwnership(nexus, oldObject)
		if err != nil {
			log.Warnf("Failed to remove ownership from old secret: %v. Has it been tampered with?", err)
		}
	}

	if !isOwner(nexus, newObject) {
		newObject.SetOwnerReferences(append(newObject.GetOwnerReferences(), ownerReference(nexus)))
		err := r.client.Update(context.TODO(), newObject)
		if err != nil {
			return fmt.Errorf("couldn't add the Operator as owner for the secret (%s): %v", newObject.GetName(), err)
		}
	}
	return nil
}

func removeOwnership(nexus *v1alpha1.Nexus, object resource.KubernetesResource) error {
	for i, ref := range object.GetOwnerReferences() {
		if nexus.UID == ref.UID {
			object.SetOwnerReferences(append(object.GetOwnerReferences()[:i], object.GetOwnerReferences()[i+1:]...))
			return nil
		}
	}
	return fmt.Errorf("nexus/%s was not an owner of secret/%s", nexus.Name, object.GetName())
}

func isOwner(nexus *v1alpha1.Nexus, object resource.KubernetesResource) bool {
	for _, ref := range object.GetOwnerReferences() {
		if nexus.UID == ref.UID {
			return true
		}
	}
	return false
}

func ownerReference(nexus *v1alpha1.Nexus) v12.OwnerReference {
	isController := true
	return v12.OwnerReference{
		UID:        nexus.UID,
		Name:       nexus.Name,
		APIVersion: nexus.APIVersion,
		Kind:       nexus.Kind,
		Controller: &isController,
	}
}
