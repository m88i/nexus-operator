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

package persistence

import (
	ctx "context"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const mgrNotInit = "the manager has not been initialized"

var log = logger.GetLogger("persistence_manager")

// manager is responsible for creating persistence resources, fetching deployed ones and comparing them
type manager struct {
	nexus  *v1alpha1.Nexus
	client client.Client
}

// NewManager creates a persistence resources manager
func NewManager(nexus *v1alpha1.Nexus, client client.Client) *manager {
	return &manager{
		nexus:  nexus,
		client: client,
	}
}

// GetRequiredResources returns the resources initialized by the manager
func (m *manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	var resources []resource.KubernetesResource
	if m.nexus.Spec.Persistence.Persistent {
		log.Debugf("Creating Persistent Volume Claim (%s)", m.nexus.Name)
		pvc := newPVC(m.nexus)
		resources = append(resources, pvc)
	}
	return resources, nil
}

// GetDeployedResources returns the persistence resources deployed on the cluster
func (m *manager) GetDeployedResources() ([]resource.KubernetesResource, error) {
	if m.nexus == nil || m.client == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}

	var resources []resource.KubernetesResource
	if pvc, err := m.getDeployedPVC(); err == nil {
		resources = append(resources, pvc)
	} else if !errors.IsNotFound(err) {
		log.Errorf("Could not fetch Persistent Volume Claim (%s): %v", m.nexus.Name, err)
		return nil, fmt.Errorf("could not fetch pvc (%s): %v", m.nexus.Name, err)
	}
	return resources, nil
}

func (m *manager) getDeployedPVC() (resource.KubernetesResource, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	log.Debugf("Attempting to fetch deployed Persistent Volume Claim (%s)", m.nexus.Name)
	err := m.client.Get(ctx.TODO(), key, pvc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed Persistent Volume Claim (%s)", m.nexus.Name)
		}
		return nil, err
	}
	return pvc, nil
}

// GetCustomComparator returns the custom comp function used to compare a persistence resource.
// Returns nil if there is none
func (m *manager) GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	// As PVCs have a default comparator we just return nil here
	return nil
}

// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
// Returns nil if there are none
func (m *manager) GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	// As PVCs have a default comparator we just return nil here
	return nil
}
