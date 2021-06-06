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
	"testing"

	v1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/meta"
)

var (
	routeNexus = &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nexus3",
			Namespace: "nexus",
		},
		Spec: v1alpha1.NexusSpec{
			Networking: v1alpha1.NexusNetworking{
				Annotations: map[string]string{
					"test-annotation": "enabled",
				},
				Expose:   true,
				ExposeAs: v1alpha1.RouteExposeType,
				Host:     "route.tls.test.com",
				TLS: v1alpha1.NexusNetworkingTLS{
					Mandatory: true,
				},
			},
		},
	}

	routeService = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nexus3",
			Namespace: "nexus",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{TargetPort: intstr.FromString(deployment.NexusPortName)},
			},
		},
	}
)

func TestNewRoute(t *testing.T) {
	route := newRouteBuilder(routeNexus).build()
	assertRouteBasic(t, route)
}

func TestNewRouteWithRedirection(t *testing.T) {
	route := newRouteBuilder(routeNexus).withRedirect().build()
	assertRouteBasic(t, route)
	assertRouteRedirection(t, route)
}

func assertRouteBasic(t *testing.T, route *v1.Route) {
	assert.Equal(t, routeNexus.Name, route.Name)
	assert.Equal(t, routeNexus.Namespace, route.Namespace)
	assert.Len(t, route.Labels, 1)
	assert.Equal(t, nexusIngress.Name, route.Labels[meta.AppLabel])
	assert.Equal(t, "enabled", route.Annotations["test-annotation"])

	assert.NotNil(t, route.Spec)

	assert.NotNil(t, route.Spec.To)
	assert.Equal(t, serviceKind, route.Spec.To.Kind)
	assert.Equal(t, routeService.Name, route.Spec.To.Name)

	assert.NotNil(t, route.Spec.Port)
	assert.Equal(t, routeService.Spec.Ports[0].TargetPort, route.Spec.Port.TargetPort)
}

func assertRouteRedirection(t *testing.T, route *v1.Route) {
	assert.NotNil(t, route.Spec.TLS)
	assert.Equal(t, v1.InsecureEdgeTerminationPolicyRedirect, route.Spec.TLS.InsecureEdgeTerminationPolicy)
	assert.Equal(t, v1.TLSTerminationEdge, route.Spec.TLS.Termination)
}
