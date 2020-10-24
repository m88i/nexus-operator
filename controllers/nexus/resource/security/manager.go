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
	"strings"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	secv1 "github.com/openshift/api/security/v1"
	core "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/validation"
	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/m88i/nexus-operator/pkg/logger"
)

// Manager is responsible for creating security resources, fetching deployed ones and comparing them
// Use with zero values will result in a panic. Use the NewManager function to get a properly initialized manager
type Manager struct {
	nexus             *v1alpha1.Nexus
	client            client.Client
	log               logger.Logger
	isOCP             bool
	managedObjectsRef map[string]resource.KubernetesResource
}

// NewManager creates a security resources Manager
func NewManager(nexus *v1alpha1.Nexus, client client.Client) (*Manager, error) {
	mgr := &Manager{
		nexus:  nexus,
		client: client,
		log:    logger.GetLoggerWithResource("security_manager", nexus),

		managedObjectsRef: map[string]resource.KubernetesResource{
			framework.SecretKind:     &core.Secret{},
			framework.SvcAccountKind: &core.ServiceAccount{},
		},
	}

	isOCP, err := discovery.IsOpenShift()
	if err != nil {
		return nil, fmt.Errorf("unable to determine if on Openshift: %v", err)
	}
	if isOCP {
		mgr.isOCP = true
		mgr.managedObjectsRef[framework.SCCKind] = &secv1.SecurityContextConstraints{}
		mgr.managedObjectsRef[framework.RoleBindingKind] = &rbacv1.RoleBinding{}
		mgr.managedObjectsRef[framework.ClusterRoleKind] = &rbacv1.ClusterRole{}
	}

	return mgr, nil
}

// GetRequiredResources returns the resources initialized by the Manager
func (m *Manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	m.log.Debug("Generating required resource", "kind", framework.SvcAccountKind)
	m.log.Debug("Generating required resource", "kind", framework.SecretKind)
	resources := []resource.KubernetesResource{defaultServiceAccount(m.nexus), defaultSecret(m.nexus)}

	if m.isOCP {
		// the SCC and the ClusterRole are cluster-scoped, but still part of the desired state
		// so we should ensure they're properly configured on each reconciliation
		m.log.Debug("Generating required resource", "kind", framework.SCCKind)
		resources = append(resources, defaultSCC())
		m.log.Debug("Generating required resource", "kind", framework.ClusterRoleKind)
		resources = append(resources, defaultClusterRole())

		// we only want to bind the Service Account to the ClusterRole if using the community image
		if m.nexus.Spec.Image == strings.Split(validation.NexusCommunityImage, ":")[0] {
			m.log.Debug("Generating required resource", "kind", framework.RoleBindingKind)
			resources = append(resources, defaultRoleBinding(m.nexus))
		}
	}

	return resources, nil
}

// GetDeployedResources returns the security resources deployed on the cluster
func (m *Manager) GetDeployedResources() ([]resource.KubernetesResource, error) {
	return framework.FetchDeployedResources(m.managedObjectsRef, m.nexus, m.client)
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
