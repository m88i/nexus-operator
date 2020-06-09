//     Copyright 2020 Nexus Operator and/or its authors
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

package infra

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"reflect"
)

// Supervisor is the resources manager for the nexus CR.
// Handles the creation of every single resource needed to deploy a nexus server instance on Kubernetes
type Supervisor interface {
	// InitManagers initializes the managers responsible for the resources life cycle
	InitManagers(nexus *v1alpha1.Nexus) error
	// GetDeployedResources will fetch for the resources managed by the nexus instance deployed in the cluster
	GetDeployedResources() (resources map[reflect.Type][]resource.KubernetesResource, err error)
	// GetRequiredResources will create the requests resources as it's supposed to be
	GetRequiredResources() (resources map[reflect.Type][]resource.KubernetesResource, err error)
	// GetComparator returns the comparator based on the active managers
	GetComparator() (compare.MapComparator, error)
}

type Manager interface {
	// GetRequiredResources returns the resources initialized by the manager
	GetRequiredResources() ([]resource.KubernetesResource, error)
	// GetDeployedResources returns the resources deployed on the cluster
	GetDeployedResources() ([]resource.KubernetesResource, error)
	// GetCustomComparator returns the custom comp function used to compare two resources of a specific type
	// Returns nil if there is no custom comparator for that type
	GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
	// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
	// Returns nil if there are none
	GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
}
