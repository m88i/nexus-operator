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
	ctx "context"
	"fmt"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	"github.com/m88i/nexus-operator/pkg/logger"
	routev1 "github.com/openshift/api/route/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	discOCPFailureFormat = "unable to determine if cluster is Openshift: %v"
	discFailureFormat    = "unable to determine if %s are available: %v" // resource type, error
	resUnavailableFormat = "%s are not available in this cluster"        // resource type
)

var log = logger.GetLogger("networking_manager")

// Manager is responsible for creating networking resources, fetching deployed ones and comparing them
// Use with zero values will result in a panic. Use the NewManager function to get a properly initialized manager
type Manager struct {
	nexus  *v1alpha1.Nexus
	client client.Client

	routeAvailable, ingressAvailable, ocp bool
}

// NewManager creates a networking resources manager
// It is expected that the Nexus has been previously validated.
func NewManager(nexus *v1alpha1.Nexus, client client.Client, disc discovery.DiscoveryInterface) (*Manager, error) {
	routeAvailable, err := openshift.IsRouteAvailable(disc)
	if err != nil {
		return nil, fmt.Errorf(discFailureFormat, "routes", err)
	}

	ingressAvailable, err := kubernetes.IsIngressAvailable(disc)
	if err != nil {
		return nil, fmt.Errorf(discFailureFormat, "ingresses", err)
	}

	ocp, err := openshift.IsOpenShift(disc)
	if err != nil {
		return nil, fmt.Errorf(discOCPFailureFormat, err)
	}

	return &Manager{
		nexus:            nexus,
		client:           client,
		routeAvailable:   routeAvailable,
		ingressAvailable: ingressAvailable,
		ocp:              ocp,
	}, nil
}

func (m *Manager) IngressAvailable() bool {
	return m.ingressAvailable
}

func (m *Manager) RouteAvailable() bool {
	return m.routeAvailable
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

		log.Debugf("Creating Route (%s)", m.nexus.Name)
		route := m.createRoute()
		resources = append(resources, route)

	case v1alpha1.IngressExposeType:
		if !m.ingressAvailable {
			return nil, fmt.Errorf(resUnavailableFormat, "ingresses")
		}

		log.Debugf("Creating Ingress (%s)", m.nexus.Name)
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

func (m *Manager) createIngress() *networkingv1beta1.Ingress {
	builder := newIngressBuilder(m.nexus)
	if len(m.nexus.Spec.Networking.TLS.SecretName) > 0 {
		builder = builder.withCustomTLS()
	}
	return builder.build()
}

// GetDeployedResources returns the networking resources deployed on the cluster
func (m *Manager) GetDeployedResources() ([]resource.KubernetesResource, error) {
	var resources []resource.KubernetesResource
	if m.routeAvailable {
		if route, err := m.getDeployedRoute(); err == nil {
			resources = append(resources, route)
		} else if !errors.IsNotFound(err) {
			log.Errorf("Could not fetch Route (%s): %v", m.nexus.Name, err)
			return nil, fmt.Errorf("could not fetch route (%s): %v", m.nexus.Name, err)
		}
	}
	if m.ingressAvailable {
		if ingress, err := m.getDeployedIngress(); err == nil {
			resources = append(resources, ingress)
		} else if !errors.IsNotFound(err) {
			log.Errorf("Could not fetch Ingress (%s): %v", m.nexus.Name, err)
			return nil, fmt.Errorf("could not fetch ingress (%s): %v", m.nexus.Name, err)
		}
	}
	return resources, nil
}

func (m *Manager) getDeployedRoute() (*routev1.Route, error) {
	route := &routev1.Route{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	err := m.client.Get(ctx.TODO(), key, route)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed Route (%s)", m.nexus.Name)
		}
		return nil, err
	}
	return route, nil
}

func (m *Manager) getDeployedIngress() (*networkingv1beta1.Ingress, error) {
	ingress := &networkingv1beta1.Ingress{}
	key := types.NamespacedName{Namespace: m.nexus.Namespace, Name: m.nexus.Name}
	err := m.client.Get(ctx.TODO(), key, ingress)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed Ingress (%s)", m.nexus.Name)
		}
		return nil, err
	}
	return ingress, nil
}

// GetCustomComparator returns the custom comp function used to compare a networking resource.
// Returns nil if there is none
func (m *Manager) GetCustomComparator(t reflect.Type) func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	if t == reflect.TypeOf(&networkingv1beta1.Ingress{}) {
		return ingressEqual
	}
	return nil
}

// GetCustomComparators returns all custom comp functions in a map indexed by the resource type
// Returns nil if there are none
func (m *Manager) GetCustomComparators() map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	ingressType := reflect.TypeOf(networkingv1beta1.Ingress{})
	return map[reflect.Type]func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool{
		ingressType: ingressEqual,
	}
}

func ingressEqual(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	ingress1 := deployed.(*networkingv1beta1.Ingress)
	ingress2 := requested.(*networkingv1beta1.Ingress)
	var pairs [][2]interface{}
	pairs = append(pairs, [2]interface{}{ingress1.Name, ingress2.Name})
	pairs = append(pairs, [2]interface{}{ingress1.Namespace, ingress2.Namespace})
	pairs = append(pairs, [2]interface{}{ingress1.Spec, ingress2.Spec})

	equal := compare.EqualPairs(pairs)
	if !equal {
		log.Info("Resources are not equal", "deployed", deployed, "requested", requested)
	}
	return equal
}
