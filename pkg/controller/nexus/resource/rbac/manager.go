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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const mgrNotInit = "the manager has not been initialized"

var log = logger.GetLogger("rbac_management")

// Manager is responsible for creating RBAC resources, fetching deployed ones and comparing them
type Manager struct {
	scvAccount *core.ServiceAccount
	nexus      *v1alpha1.Nexus
	client     client.Client
}

// NewDefaultManager creates a manager using all default resources
func NewDefaultManager(nexus *v1alpha1.Nexus, client client.Client) *Manager {
	return (&Manager{}).withAllDefaults(nexus).withClient(client)
}

func (m *Manager) withAllDefaults(nexus *v1alpha1.Nexus) *Manager {
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
	if m.scvAccount == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}
	return map[reflect.Type][]resource.KubernetesResource{
		reflect.TypeOf(m.scvAccount): {m.scvAccount},
	}, nil
}

// GetDeployedResources returns the rbac resources deployed on the cluster
func (m *Manager) GetDeployedResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	if m.nexus == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}
	if m.client == nil {
		return nil, fmt.Errorf("the client has not been initialized")
	}

	resources := make(map[reflect.Type][]resource.KubernetesResource)
	svcAccnt, err := m.getDeployedSvcAccnt()
	if err == nil {
		resources[reflect.TypeOf(svcAccnt)] = []resource.KubernetesResource{svcAccnt}
	} else if !errors.IsNotFound(err) {
		return nil, fmt.Errorf("couldn't fetch ServiceAccount (%s): %v", m.nexus.Name, err)
	}

	return resources, nil
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

// GetComparator returns the function used to compare a rbac resource. Can take a ServiceAccount as argument
func GetComparator(p reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	if p == reflect.TypeOf(core.ServiceAccount{}) {
		return svcAccntEqual
	}
	log.Errorf("argument resource is not a ServiceAccount")
	return nil
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
