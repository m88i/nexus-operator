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

package deployment

import (
	"fmt"
	"reflect"

	"strings"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var managedObjectsRef = map[string]resource.KubernetesResource{
	"Service":    &corev1.Service{},
	"Deployment": &appsv1.Deployment{},
}

// Manager is responsible for creating deployment-related resources, fetching deployed ones and comparing them
// Use with zero values will result in a panic. Use the NewManager function to get a properly initialized manager
type Manager struct {
	nexus  *v1alpha1.Nexus
	client client.Client
}

// NewManager creates a deployment resources manager
// It is expected that the Nexus has been previously validated.
func NewManager(nexus *v1alpha1.Nexus, client client.Client) *Manager {
	return &Manager{
		nexus:  nexus,
		client: client,
	}
}

// GetRequiredResources returns the resources initialized by the manager
func (m *Manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	return []resource.KubernetesResource{newDeployment(m.nexus), newService(m.nexus)}, nil
}

// GetDeployedResources returns the deployment-related resources deployed on the cluster
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

// GetCustomComparator returns the custom comp function used to compare a deployment-related resource
// Returns nil if there is none
func (m *Manager) GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	if t == reflect.TypeOf(&appsv1.Deployment{}) {
		return deploymentEqual
	}
	return nil
}

// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
// Returns nil if there are none
func (m *Manager) GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	deploymentType := reflect.TypeOf(appsv1.Deployment{})
	return map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool{
		deploymentType: deploymentEqual,
	}
}

func deploymentEqual(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	depDeployment := deployed.(*appsv1.Deployment)
	reqDeployment := requested.(*appsv1.Deployment)

	normalizeSecurityContext(depDeployment, reqDeployment)

	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{depDeployment.Name, reqDeployment.Name})
	pairs = append(pairs, [2]interface{}{depDeployment.Namespace, reqDeployment.Namespace})
	pairs = append(pairs, [2]interface{}{depDeployment.Labels, reqDeployment.Labels})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Replicas, reqDeployment.Spec.Replicas})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Selector, reqDeployment.Spec.Selector})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.ObjectMeta, reqDeployment.Spec.Template.ObjectMeta})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.Volumes, reqDeployment.Spec.Template.Spec.Volumes})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.ServiceAccountName, reqDeployment.Spec.Template.Spec.ServiceAccountName})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.SecurityContext, reqDeployment.Spec.Template.Spec.SecurityContext})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.Containers[0].Name, reqDeployment.Spec.Template.Spec.Containers[0].Name})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.Containers[0].Ports, reqDeployment.Spec.Template.Spec.Containers[0].Ports})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.Containers[0].Resources, reqDeployment.Spec.Template.Spec.Containers[0].Resources})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.Containers[0].Image, reqDeployment.Spec.Template.Spec.Containers[0].Image})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.Containers[0].LivenessProbe, reqDeployment.Spec.Template.Spec.Containers[0].LivenessProbe})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe, reqDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe})
	pairs = append(pairs, [2]interface{}{depDeployment.Spec.Template.Spec.Containers[0].Env, reqDeployment.Spec.Template.Spec.Containers[0].Env})

	equal := compare.EqualPairs(pairs)
	equal = equal && equalPullPolicies(depDeployment, reqDeployment)
	return equal
}

func equalPullPolicies(depDeployment, reqDeployment *appsv1.Deployment) bool {
	if len(reqDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy) > 0 {
		return reqDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy == depDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy
	}

	reqImageParts := strings.Split(reqDeployment.Spec.Template.Spec.Containers[0].Image, ":")
	if len(reqImageParts) == 1 || reqImageParts[1] == "latest" {
		return depDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy == corev1.PullAlways
	}

	return depDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy == corev1.PullIfNotPresent
}

// see: https://github.com/m88i/nexus-operator/issues/156
// On OpenShift 4.5+ `SecurityContext` is not nil, but a "blank" object.
// Since we are requesting a nil object in this context, we consider the deployed object to be nil as well.
func normalizeSecurityContext(depDeployment, reqDeployment *appsv1.Deployment) {
	if reqDeployment.Spec.Template.Spec.SecurityContext == nil {
		depDeployment.Spec.Template.Spec.SecurityContext = nil
	}
}
