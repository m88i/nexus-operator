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

package security

import (
	"fmt"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/m88i/nexus-operator/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/framework"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var managedObjectsRef = map[string]resource.KubernetesResource{
	"Secret":          &core.Secret{},
	"Service Account": &core.ServiceAccount{},
}

// Manager is responsible for creating security resources, fetching deployed ones and comparing them
// Use with zero values will result in a panic. Use the NewManager function to get a properly initialized manager
type Manager struct {
	nexus  *v1alpha1.Nexus
	client client.Client
}

// NewManager creates a security resources Manager
func NewManager(nexus *v1alpha1.Nexus, client client.Client) *Manager {
	return &Manager{
		nexus:  nexus,
		client: client,
	}
}

// GetRequiredResources returns the resources initialized by the Manager
func (m *Manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	return []resource.KubernetesResource{defaultServiceAccount(m.nexus), defaultSecret(m.nexus)}, nil
}

// GetDeployedResources returns the security resources deployed on the cluster
func (m *Manager) GetDeployedResources() ([]resource.KubernetesResource, error) {
	var resources []resource.KubernetesResource
	for resType, resRef := range managedObjectsRef {
		if err := framework.Fetch(m.client, framework.Key(m.nexus), resRef); err == nil {
			resources = append(resources, resRef)
		} else if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("could not fetch Resource %s (%s): %v", resType, m.nexus.Name, err)
		}
	}
	return resources, nil
}

// GetCustomComparator returns the custom comp function used to compare a security resource.
// Returns nil if there is none
func (m *Manager) GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	if t == reflect.TypeOf(&core.Secret{}) {
		return framework.AlwaysTrueComparator()
	}
	return nil
}

// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
// Returns nil if there are none
func (m *Manager) GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool{
		reflect.TypeOf(core.Secret{}): framework.AlwaysTrueComparator(),
	}
}
