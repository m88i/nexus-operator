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

package security

import (
	ctx "context"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const mgrNotInit = "the manager has not been initialized"

var log = logger.GetLogger("security_manager")

// Manager is responsible for creating security resources, fetching deployed ones and comparing them
type Manager struct {
	nexus  *v1alpha1.Nexus
	client client.Client
}

// NewManager creates a security resources Manager
func NewManager(nexus v1alpha1.Nexus, client client.Client) *Manager {
	mgr := &Manager{
		nexus:  &nexus,
		client: client,
	}
	mgr.setDefaults()
	return mgr
}

// setDefaults destructively sets default for unset values in the Nexus CR
func (m *Manager) setDefaults() {
	if len(m.nexus.Spec.ServiceAccountName) == 0 {
		m.nexus.Spec.ServiceAccountName = m.nexus.Name
	}
}

// GetRequiredResources returns the resources initialized by the Manager
func (m *Manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	if m.nexus == nil || m.client == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}

	log.Debugf("Creating Service Account (%s)", m.nexus.Name)
	svcAccnt := defaultServiceAccount(m.nexus)
	return []resource.KubernetesResource{svcAccnt}, nil
}

// GetDeployedResources returns the security resources deployed on the cluster
func (m *Manager) GetDeployedResources() ([]resource.KubernetesResource, error) {
	if m.nexus == nil || m.client == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}

	var resources []resource.KubernetesResource
	if svcAccnt, err := m.getDeployedSvcAccnt(); err == nil {
		resources = append(resources, svcAccnt)
	} else if !errors.IsNotFound(err) {
		log.Errorf("Could not fetch Service Account (%s): %v", m.nexus.Name, err)
		return nil, fmt.Errorf("could not fetch service account (%s): %v", m.nexus.Name, err)
	}
	return resources, nil
}

func (m *Manager) getDeployedSvcAccnt() (resource.KubernetesResource, error) {
	account := &core.ServiceAccount{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	log.Debugf("Attempting to fetch deployed Service Account (%s)", m.nexus.Name)
	err := m.client.Get(ctx.TODO(), key, account)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed Service Account (%s)", m.nexus.Name)
		}
		return nil, err
	}
	return account, nil
}

// GetCustomComparator returns the custom comp function used to compare a security resource.
// Returns nil if there is none
func (m *Manager) GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	// As Service Accounts have a default comparator we just return nil here
	return nil
}

// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
// Returns nil if there are none
func (m *Manager) GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	// As Service Accounts have a default comparator we just return nil here
	return nil
}
