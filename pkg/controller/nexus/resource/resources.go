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
	"fmt"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/infra"

	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/networking"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/persistence"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/security"
	"reflect"

	"k8s.io/client-go/discovery"

	"github.com/RHsyseng/operator-utils/pkg/resource"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

const mgrsNotInit = "resource managers have not been initialized"

type supervisor struct {
	client          client.Client
	discoveryClient discovery.DiscoveryInterface
	managers        []infra.Manager
}

// NewSupervisor creates a new resource manager for nexus CR
func NewSupervisor(client client.Client, discoveryClient discovery.DiscoveryInterface) infra.Supervisor {
	return &supervisor{
		client:          client,
		discoveryClient: discoveryClient,
	}
}

// InitManagers initializes the managers responsible for the resources life cycle
func (r *supervisor) InitManagers(nexus *v1alpha1.Nexus) error {
	networkManager, err := networking.NewManager(nexus, r.client, r.discoveryClient)
	if err != nil {
		return fmt.Errorf("unable to create networking manager: %v", err)
	}

	r.managers = []infra.Manager{
		deployment.NewManager(nexus, r.client),
		persistence.NewManager(nexus, r.client),
		security.NewManager(nexus, r.client),
		networkManager,
	}
	return nil
}

func (r *supervisor) GetDeployedResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	if len(r.managers) == 0 {
		return nil, fmt.Errorf(mgrsNotInit)
	}

	log.Info("Fetching deployed resources")
	builder := compare.NewMapBuilder()
	for _, manager := range r.managers {
		deployedResources, err := manager.GetDeployedResources()
		if err != nil {
			return nil, err
		}
		builder.Add(deployedResources...)
	}
	return builder.ResourceMap(), nil
}

func (r *supervisor) GetRequiredResources() (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	if len(r.managers) == 0 {
		return nil, fmt.Errorf(mgrsNotInit)
	}

	log.Info("Fetching required resources")
	builder := compare.NewMapBuilder()
	for _, manager := range r.managers {
		requiredResources, err := manager.GetRequiredResources()
		if err != nil {
			return nil, err
		}
		builder.Add(requiredResources...)
	}
	return builder.ResourceMap(), nil
}

// GetComparator will create the comparator for the Nexus instance
// The comparator can be used to compare two different sets of resources and update them accordingly
func (r *supervisor) GetComparator() (compare.MapComparator, error) {
	if len(r.managers) == 0 {
		return compare.MapComparator{}, fmt.Errorf(mgrsNotInit)
	}

	resourceComparator := compare.DefaultComparator()
	for _, manager := range r.managers {
		for resType, compFunc := range manager.GetCustomComparators() {
			resourceComparator.SetComparator(resType, compFunc)
		}
	}
	return compare.MapComparator{Comparator: resourceComparator}, nil
}
