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

package deployment

import (
	ctx "context"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var log = logger.GetLogger("deployment_manager")

// Manager is responsible for creating deployment-related resources, fetching deployed ones and comparing them
// Use with zero values will result in a panic. Use the NewManager function to get a properly initialized manager
type Manager struct {
	nexus  *v1alpha1.Nexus
	client client.Client
}

// NewManager creates a deployment resources manager.
// It is expected that the Nexus has been previously validated.
func NewManager(nexus v1alpha1.Nexus, client client.Client) *Manager {
	return &Manager{
		nexus:  &nexus,
		client: client,
	}
}

// GetRequiredResources returns the resources initialized by the manager
func (m *Manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	log.Debugf("Creating Deployment (%s)", m.nexus.Name)
	deployment := newDeployment(m.nexus)
	log.Debugf("Creating Service (%s)", m.nexus.Name)
	svc := newService(m.nexus)
	return []resource.KubernetesResource{deployment, svc}, nil
}

// GetDeployedResources returns the deployment-related resources deployed on the cluster
func (m *Manager) GetDeployedResources() ([]resource.KubernetesResource, error) {
	var resources []resource.KubernetesResource
	if deployment, err := m.getDeployedDeployment(); err == nil {
		resources = append(resources, deployment)
	} else if !errors.IsNotFound(err) {
		log.Errorf("Could not fetch Deployment (%s): %v", m.nexus.Name, err)
		return nil, fmt.Errorf("could not fetch service (%s): %v", m.nexus.Name, err)
	}
	if service, err := m.getDeployedService(); err == nil {
		resources = append(resources, service)
	} else if !errors.IsNotFound(err) {
		log.Errorf("Could not fetch Service (%s): %v", m.nexus.Name, err)
		return nil, fmt.Errorf("could not fetch service (%s): %v", m.nexus.Name, err)
	}
	return resources, nil
}

func (m *Manager) getDeployedDeployment() (resource.KubernetesResource, error) {
	dep := &appsv1.Deployment{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	log.Debugf("Attempting to fetch deployed Deployment (%s)", m.nexus.Name)
	err := m.client.Get(ctx.TODO(), key, dep)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed Service (%s)", m.nexus.Name)
		}
		return nil, err
	}
	return dep, nil
}

func (m *Manager) getDeployedService() (resource.KubernetesResource, error) {
	svc := &corev1.Service{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	log.Debugf("Attempting to fetch deployed Service (%s)", m.nexus.Name)
	err := m.client.Get(ctx.TODO(), key, svc)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed Service (%s)", m.nexus.Name)
		}
		return nil, err
	}
	return svc, nil
}

// GetCustomComparator returns the custom comp function used to compare a deployment-related resource
// Returns nil if there is none
func (m *Manager) GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	// As Deployments and Services have default comparators we just return nil here
	return nil
}

// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
// Returns nil if there are none
func (m *Manager) GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	// As Deployments and Services have default comparators we just return nil here
	return nil
}
