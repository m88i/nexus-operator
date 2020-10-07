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

package resource

import (
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
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
