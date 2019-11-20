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
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/client-go/discovery"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

var log = logf.Log.WithName("controller_nexus")

// NexusResourceManager is the resources manager for the Nexus CR.
// Handles the creation of every single resource needed to deploy a Nexus server instance on Kubernetes
type NexusResourceManager interface {
	// GetDeployedResources will fetch for the resources managed by the Nexus instance deployed in the cluster
	GetDeployedResources(nexus *v1alpha1.Nexus) (resources map[reflect.Type][]resource.KubernetesResource, err error)
	// CreateRequiredResources will create the requests resources as it's supposed to be
	CreateRequiredResources(nexus *v1alpha1.Nexus) (resources map[reflect.Type][]resource.KubernetesResource, err error)
}

type nexusResourceManager struct {
	NexusResourceManager
	client          client.Client
	discoveryClient discovery.DiscoveryInterface
}

// New creates a new resource manager for Nexus CR
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
		resources, err = reader.ListAll(&v1.PersistentVolumeClaimList{}, &v1.ServiceList{}, &appsv1.DeploymentList{}, &v1beta1.IngressList{}, &routev1.RouteList{})
	} else {
		resources, err = reader.ListAll(&v1.PersistentVolumeClaimList{}, &v1.ServiceList{}, &appsv1.DeploymentList{}, &v1beta1.IngressList{})
	}

	if err != nil {
		log.Error(err, "Failed to fetch deployed Nexus resources")
		return nil, err
	}
	if resources == nil {
		log.Info("No deployed resources found")
	}
	log.Info("Number of deployed ", "resources", len(resources))
	return resources, nil
}

func (r *nexusResourceManager) CreateRequiredResources(nexus *v1alpha1.Nexus) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	logger := log.WithValues("Nexus.Namespace", nexus.Namespace, "Nexus.Name", nexus.Name)
	logger.Info("Creating resources structures")
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
			if available, err := openshift.IsRouteAvailable(r.discoveryClient); err != nil {
				return nil, err
			} else if available {
				route := newRoute(nexus, service)
				resources[reflect.TypeOf(routev1.Route{})] = []resource.KubernetesResource{route}
			}
		case v1alpha1.IngressExposeType:
			ingress, err := newIngress(nexus, service)
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
