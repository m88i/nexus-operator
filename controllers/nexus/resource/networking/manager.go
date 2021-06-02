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

package networking

import (
	"fmt"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/m88i/nexus-operator/pkg/framework/kind"
	"github.com/m88i/nexus-operator/pkg/logger"
)

const (
	discFailureFormat    = "unable to determine if %s are available: %v" // resource type, error
	resUnavailableFormat = "%s are not available in this cluster"        // resource type
)

var (
	legacyIngressType = reflect.TypeOf(&networkingv1beta1.Ingress{})
	ingressType       = reflect.TypeOf(&networkingv1.Ingress{})
)

// Manager is responsible for creating networking resources, fetching deployed ones and comparing them
// Use with zero values will result in a panic. Use the NewManager function to get a properly initialized manager
type Manager struct {
	nexus             *v1alpha1.Nexus
	client            client.Client
	log               logger.Logger
	managedObjectsRef map[string]resource.KubernetesResource

	routeAvailable, ingressAvailable, legacyIngressAvailable bool
}

// NewManager creates a networking resources manager
// It is expected that the Nexus has been previously validated.
func NewManager(nexus *v1alpha1.Nexus, client client.Client) (*Manager, error) {
	mgr := &Manager{
		nexus:             nexus,
		client:            client,
		log:               logger.GetLoggerWithResource("networking_manager", nexus),
		managedObjectsRef: make(map[string]resource.KubernetesResource),
	}

	routeAvailable, err := discovery.IsRouteAvailable()
	if err != nil {
		return nil, fmt.Errorf(discFailureFormat, "routes", err)
	}

	ingressAvailable, err := discovery.IsIngressAvailable()
	if err != nil {
		return nil, fmt.Errorf(discFailureFormat, "ingresses", err)
	}

	legacyIngressAvailable, err := discovery.IsLegacyIngressAvailable()
	if err != nil {
		return nil, fmt.Errorf(discFailureFormat, "v1beta1 ingresses", err)
	}

	if ingressAvailable {
		mgr.ingressAvailable = true
		mgr.managedObjectsRef[kind.IngressKind] = &networkingv1.Ingress{}
	} else if legacyIngressAvailable {
		mgr.legacyIngressAvailable = true
		mgr.managedObjectsRef[kind.IngressKind] = &networkingv1beta1.Ingress{}
	}

	if routeAvailable {
		mgr.routeAvailable = true
		mgr.managedObjectsRef[kind.RouteKind] = &routev1.Route{}
	}

	return mgr, nil
}

// GetRequiredResources returns the resources initialized by the manager
func (m *Manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	if !m.nexus.Spec.Networking.Expose {
		return nil, nil
	}

	var resources []resource.KubernetesResource
	switch m.nexus.Spec.Networking.ExposeAs {
	case v1alpha1.RouteExposeType:
		if !m.routeAvailable {
			return nil, fmt.Errorf(resUnavailableFormat, "routes")
		}

		m.log.Debug("Generating required resource", "kind", kind.RouteKind)
		route := m.createRoute()
		resources = append(resources, route)

	case v1alpha1.IngressExposeType:
		if !m.ingressAvailable && !m.legacyIngressAvailable {
			return nil, fmt.Errorf(resUnavailableFormat, "ingresses")
		}

		m.log.Debug("Generating required resource", "kind", kind.IngressKind)
		ingress := m.createIngress()
		resources = append(resources, ingress)
	}
	return resources, nil
}

func (m *Manager) createRoute() *routev1.Route {
	builder := newRouteBuilder(m.nexus)
	if m.nexus.Spec.Networking.TLS.Mandatory {
		builder = builder.withRedirect()
	}
	return builder.build()
}

func (m *Manager) createIngress() resource.KubernetesResource {
	// we're only here if either ingress is available, no need to check legacy is
	if !m.ingressAvailable {
		builder := newLegacyIngressBuilder(m.nexus)
		if len(m.nexus.Spec.Networking.TLS.SecretName) > 0 {
			builder = builder.withCustomTLS()
		}
		return builder.build()
	}
	builder := newIngressBuilder(m.nexus)
	if len(m.nexus.Spec.Networking.TLS.SecretName) > 0 {
		builder = builder.withCustomTLS()
	}
	return builder.build()
}

// GetDeployedResources returns the networking resources deployed on the cluster
func (m *Manager) GetDeployedResources() ([]resource.KubernetesResource, error) {
	return framework.FetchDeployedResources(m.managedObjectsRef, m.nexus, m.client)
}

// GetCustomComparator returns the custom comp function used to compare a networking resource.
// Returns nil if there is none
func (m *Manager) GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	switch t {
	case legacyIngressType:
		return legacyIngressEqual
	case ingressType:
		return ingressEqual
	default:
		return nil
	}
}

// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
// Returns nil if there are none
func (m *Manager) GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool{
		legacyIngressType: legacyIngressEqual,
		ingressType:       ingressEqual,
	}
}

func legacyIngressEqual(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	ingress1 := deployed.(*networkingv1beta1.Ingress)
	ingress2 := requested.(*networkingv1beta1.Ingress)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{ingress1.Name, ingress2.Name})
	pairs = append(pairs, [2]interface{}{ingress1.Namespace, ingress2.Namespace})
	pairs = append(pairs, [2]interface{}{ingress1.Spec, ingress2.Spec})
	pairs = append(pairs, [2]interface{}{ingress1.Annotations, ingress2.Annotations})

	equal := compare.EqualPairs(pairs)
	if !equal {
		logger.GetLogger("networking_manager").Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}

func ingressEqual(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	ingress1 := deployed.(*networkingv1.Ingress)
	ingress2 := requested.(*networkingv1.Ingress)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{ingress1.Name, ingress2.Name})
	pairs = append(pairs, [2]interface{}{ingress1.Namespace, ingress2.Namespace})
	pairs = append(pairs, [2]interface{}{ingress1.Spec, ingress2.Spec})
	pairs = append(pairs, [2]interface{}{ingress1.Annotations, ingress2.Annotations})

	equal := compare.EqualPairs(pairs)
	if !equal {
		logger.GetLogger("networking_manager").Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}
