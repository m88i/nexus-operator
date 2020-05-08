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

package rbac

import (
	ctx "context"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
	core "k8s.io/api/core/v1"
	"k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = logger.GetLogger("rbac_management")

// Manager is responsible for creating RBAC resources, fetching deployed ones and comparing them
type Manager struct {
	role        *v1.Role
	roleBinding *v1.RoleBinding
	scvAccount  *core.ServiceAccount
	nexus       *v1alpha1.Nexus
	client      client.Client
}

// NewDefaultManager creates a manager using all default resources
func NewDefaultManager(nexus *v1alpha1.Nexus, client client.Client) *Manager {
	return (&Manager{}).withAllDefaults(nexus).withClient(client)
}

func (m *Manager) withAllDefaults(nexus *v1alpha1.Nexus) *Manager {
	m.role = defaultRole(nexus)
	m.roleBinding = defaultRoleBinding(nexus)
	m.scvAccount = defaultServiceAccount(nexus)
	m.nexus = nexus
	return m
}

func (m *Manager) withClient(client client.Client) *Manager {
	m.client = client
	return m
}

// GetRequiredResources returns the resources initialized by the manager
func (m *Manager) GetRequiredResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	if m.role == nil || m.roleBinding == nil || m.scvAccount == nil {
		return nil, fmt.Errorf("all resources must have been previously initialized")
	}
	return map[reflect.Type][]resource.KubernetesResource{
		reflect.TypeOf(m.role):        {m.role},
		reflect.TypeOf(m.roleBinding): {m.roleBinding},
		reflect.TypeOf(m.scvAccount):  {m.scvAccount},
	}, nil
}

// GetDeployedResources returns the rbac resources deployed on the cluster
func (m *Manager) GetDeployedResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	if m.nexus == nil {
		return nil, fmt.Errorf("the manager has not been initialized")
	}
	if m.client == nil {
		return nil, fmt.Errorf("the client has not been initialized")
	}

	resources := make(map[reflect.Type][]resource.KubernetesResource)

	role, err := m.getDeployedRole()
	if err == nil {
		resources[reflect.TypeOf(role)] = []resource.KubernetesResource{role}
	} else if !errors.IsNotFound(err) {
		return nil, fmt.Errorf("couldn't fetch Role (%s): %v", m.nexus.Name, err)
	}

	rb, err := m.getDeployedRoleBinding()
	if err == nil {
		resources[reflect.TypeOf(rb)] = []resource.KubernetesResource{rb}
	} else if !errors.IsNotFound(err) {
		return nil, fmt.Errorf("couldn't fetch RoleBinding (%s): %v", m.nexus.Name, err)
	}

	svcAccnt, err := m.getDeployedSvcAccnt()
	if err == nil {
		resources[reflect.TypeOf(svcAccnt)] = []resource.KubernetesResource{svcAccnt}
	} else if !errors.IsNotFound(err) {
		return nil, fmt.Errorf("couldn't fetch ServiceAccount (%s): %v", m.nexus.Name, err)
	}

	return resources, nil
}

func (m *Manager) getDeployedRole() (resource.KubernetesResource, error) {
	role := &v1.Role{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	err := m.client.Get(ctx.TODO(), key, role)
	if err != nil {
		return nil, err
	}
	return role, err
}

func (m *Manager) getDeployedRoleBinding() (resource.KubernetesResource, error) {
	rb := &v1.RoleBinding{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	err := m.client.Get(ctx.TODO(), key, rb)
	if err != nil {
		return nil, err
	}
	return rb, err
}

func (m *Manager) getDeployedSvcAccnt() (resource.KubernetesResource, error) {
	account := &core.ServiceAccount{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	err := m.client.Get(ctx.TODO(), key, account)
	if err != nil {
		return nil, err
	}
	return account, err
}

// GetComparator returns the function used to compare a rbac resource. Can tak a Role, RoleBinding or ServiceAccount as argument
func GetComparator(p reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	switch p {
	case reflect.TypeOf(v1.Role{}):
		return roleEqual
	case reflect.TypeOf(v1.RoleBinding{}):
		return roleBindingEqual
	case reflect.TypeOf(core.ServiceAccount{}):
		return svcAccntEqual
	default:
		log.Errorf("argument resource is not a Role, RoleBinding or ServiceAccount")
		return nil
	}
}

func roleEqual(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	role1 := deployed.(*v1.Role)
	role2 := requested.(*v1.Role)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{role1.Name, role2.Name})
	pairs = append(pairs, [2]interface{}{role1.Namespace, role2.Namespace})
	pairs = append(pairs, [2]interface{}{role1.Rules, role2.Rules})

	equal := compare.EqualPairs(pairs)
	if !equal {
		log.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func roleBindingEqual(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	rb1 := deployed.(*v1.RoleBinding)
	rb2 := requested.(*v1.RoleBinding)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{rb1.Name, rb2.Name})
	pairs = append(pairs, [2]interface{}{rb1.Namespace, rb2.Namespace})
	pairs = append(pairs, [2]interface{}{rb1.RoleRef, rb2.RoleRef})
	pairs = append(pairs, [2]interface{}{rb1.Subjects, rb2.Subjects})

	equal := compare.EqualPairs(pairs)
	if !equal {
		log.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func svcAccntEqual(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	account1 := deployed.(*core.ServiceAccount)
	account2 := requested.(*core.ServiceAccount)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{account1.Name, account2.Name})
	pairs = append(pairs, [2]interface{}{account1.Namespace, account2.Namespace})

	equal := compare.EqualPairs(pairs)
	if !equal {
		log.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}
