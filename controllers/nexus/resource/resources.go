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
	"fmt"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/networking"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/persistence"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/security"
	"github.com/m88i/nexus-operator/pkg/logger"
)

const mgrsNotInit = "resource managers have not been initialized"

type supervisor struct {
	client   client.Client
	managers []Manager
	log      logr.Logger
}

// NewSupervisor creates a new resource manager for nexus CR
func NewSupervisor(client client.Client) Supervisor {
	return &supervisor{
		client: client,
	}
}

// InitManagers initializes the managers responsible for the resources life cycle
func (s *supervisor) InitManagers(nexus *v1alpha1.Nexus) error {
	s.log = logger.GetLoggerWithResource("resource_supervisor", nexus)
	networkManager, err := networking.NewManager(nexus, s.client)
	if err != nil {
		return fmt.Errorf("unable to create networking manager: %v", err)
	}

	s.managers = []Manager{
		deployment.NewManager(nexus, s.client),
		persistence.NewManager(nexus, s.client),
		security.NewManager(nexus, s.client),
		networkManager,
	}
	return nil
}

func (s *supervisor) GetDeployedResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	if len(s.managers) == 0 {
		return nil, fmt.Errorf(mgrsNotInit)
	}

	s.log.Info("Fetching deployed resources")
	builder := compare.NewMapBuilder()
	for _, manager := range s.managers {
		deployedResources, err := manager.GetDeployedResources()
		if err != nil {
			return nil, err
		}
		builder.Add(deployedResources...)
	}
	return builder.ResourceMap(), nil
}

func (s *supervisor) GetRequiredResources() (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	if len(s.managers) == 0 {
		return nil, fmt.Errorf(mgrsNotInit)
	}

	s.log.Info("Generating required resources")
	builder := compare.NewMapBuilder()
	for _, manager := range s.managers {
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
func (s *supervisor) GetComparator() (compare.MapComparator, error) {
	if len(s.managers) == 0 {
		return compare.MapComparator{}, fmt.Errorf(mgrsNotInit)
	}

	resourceComparator := compare.DefaultComparator()
	for _, manager := range s.managers {
		for resType, compFunc := range manager.GetCustomComparators() {
			resourceComparator.SetComparator(resType, compFunc)
		}
	}
	return compare.MapComparator{Comparator: resourceComparator}, nil
}
