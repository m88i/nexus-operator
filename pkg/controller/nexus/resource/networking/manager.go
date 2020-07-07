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

package networking

import (
	ctx "context"
	"fmt"
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
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	discOCPFailureFormat      = "unable to determine if cluster is Openshift: %v"
	discFailureFormat         = "unable to determine if %s are available: %v" // resource type, error
	resUnavailableFormat      = "%s are not available in this cluster"        // resource type
	mgrNotInit                = "the manager has not been initialized"
	unspecifiedExposeAsFormat = "'spec.exposeAs' left unspecified, setting it to %s"
)

var log = logger.GetLogger("networking_manager")

// Manager is responsible for creating networking resources, fetching deployed ones and comparing them
type Manager struct {
	nexus  *v1alpha1.Nexus
	client client.Client

	routeAvailable, ingressAvailable, ocp bool
}

// NewManager creates a networking resources manager
func NewManager(nexus v1alpha1.Nexus, client client.Client, disc discovery.DiscoveryInterface) (*Manager, error) {
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

	mgr := &Manager{
		nexus:            &nexus,
		client:           client,
		routeAvailable:   routeAvailable,
		ingressAvailable: ingressAvailable,
		ocp:              ocp,
	}

	mgr.setDefaults()
	if err := mgr.validate(); err != nil {
		return nil, fmt.Errorf("unable to validate provided CR: %v", err)
	}
	return mgr, nil
}

func (m *Manager) IngressAvailable() (bool, error) {
	if m.nexus == nil || m.client == nil {
		return false, fmt.Errorf(mgrNotInit)
	}
	return m.ingressAvailable, nil
}

func (m *Manager) RouteAvailable() (bool, error) {
	if m.nexus == nil || m.client == nil {
		return false, fmt.Errorf(mgrNotInit)
	}
	return m.routeAvailable, nil
}

// setDefaults destructively sets default for unset values in the Nexus CR
func (m *Manager) setDefaults() {
	if !m.nexus.Spec.Networking.Expose {
		return
	}

	if len(m.nexus.Spec.Networking.ExposeAs) == 0 {
		if m.ocp {
			log.Infof(unspecifiedExposeAsFormat, v1alpha1.RouteExposeType)
			m.nexus.Spec.Networking.ExposeAs = v1alpha1.RouteExposeType
		} else if m.ingressAvailable {
			log.Infof(unspecifiedExposeAsFormat, v1alpha1.IngressExposeType)
			m.nexus.Spec.Networking.ExposeAs = v1alpha1.IngressExposeType
		} else {
			// we're on kubernetes < 1.14
			// try setting nodePort, validation will catch it if impossible
			log.Info("On Kubernetes, but Ingresses are not available")
			log.Infof(unspecifiedExposeAsFormat, v1alpha1.NodePortExposeType)
			m.nexus.Spec.Networking.ExposeAs = v1alpha1.NodePortExposeType
		}
	}
}

// validate checks if the networking parameters from the Nexus CR are sane
func (m *Manager) validate() error {
	if !m.nexus.Spec.Networking.Expose {
		log.Debugf("'spec.networking.expose' set to 'false', ignoring networking configuration")
		return nil
	}

	if !m.ingressAvailable && m.nexus.Spec.Networking.ExposeAs == v1alpha1.IngressExposeType {
		log.Errorf("Ingresses are not available on your cluster. Make sure to be running Kubernetes > 1.14 or if you're running Openshift set 'spec.networking.exposeAs' to '%s'. Alternatively you may also try '%s'", v1alpha1.IngressExposeType, v1alpha1.NodePortExposeType)
		return fmt.Errorf("ingress expose required, but unavailable")
	}

	if !m.routeAvailable && m.nexus.Spec.Networking.ExposeAs == v1alpha1.RouteExposeType {
		log.Errorf("Routes are not available on your cluster. If you're running Kubernetes 1.14 or higher try setting 'spec.networking.exposeAs' to '%s'. Alternatively you may also try '%s'", v1alpha1.IngressExposeType, v1alpha1.NodePortExposeType)
		return fmt.Errorf("route expose required, but unavailable")
	}

	if m.nexus.Spec.Networking.ExposeAs == v1alpha1.NodePortExposeType && m.nexus.Spec.Networking.NodePort == 0 {
		log.Errorf("NodePort networking requires a port. Check the Nexus resource 'spec.networking.nodePort' parameter")
		return fmt.Errorf("nodeport expose required, but no port informed")
	}

	if m.nexus.Spec.Networking.ExposeAs == v1alpha1.IngressExposeType && len(m.nexus.Spec.Networking.Host) == 0 {
		log.Errorf("Ingress networking requires a host. Check the Nexus resource 'spec.networking.host' parameter")
		return fmt.Errorf("ingress expose required, but no host informed")
	}

	if len(m.nexus.Spec.Networking.TLS.SecretName) > 0 && m.nexus.Spec.Networking.ExposeAs != v1alpha1.IngressExposeType {
		log.Errorf("'spec.networking.tls.secretName' is only available when using an Ingress. Try setting 'spec.networking.exposeAs' to '%s'", v1alpha1.IngressExposeType)
		return fmt.Errorf("tls secret name informed, but using route")
	}

	if m.nexus.Spec.Networking.TLS.Mandatory && m.nexus.Spec.Networking.ExposeAs != v1alpha1.RouteExposeType {
		log.Errorf("'spec.networking.tls.mandatory' is only available when using a Route. Try setting 'spec.networking.exposeAs' to '%s'", v1alpha1.RouteExposeType)
		return fmt.Errorf("tls set to mandatory, but using ingress")
	}

	return nil
}

// GetRequiredResources returns the resources initialized by the manager
func (m *Manager) GetRequiredResources() ([]resource.KubernetesResource, error) {
	if m.nexus == nil || m.client == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}
	if !m.nexus.Spec.Networking.Expose {
		return nil, nil
	}

	var resources []resource.KubernetesResource
	switch m.nexus.Spec.Networking.ExposeAs {
	case v1alpha1.RouteExposeType:
		if !m.routeAvailable {
			return nil, fmt.Errorf(resUnavailableFormat, "Routes")
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
	if m.nexus == nil || m.client == nil {
		return nil, fmt.Errorf(mgrNotInit)
	}

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
	if t == reflect.TypeOf(networkingv1beta1.Ingress{}) {
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
