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

package resource

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

var (
	routeNexus = &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nexus3",
			Namespace: "nexus",
		},
		Spec: v1alpha1.NexusSpec{
			Networking: v1alpha1.NexusNetworking{
				Host: "route.tls.test.com",
				TLS: v1alpha1.NexusNetworkingTLS{
					Mandatory: true,
				},
			},
		},
	}

	routeService = &corev1.Service{
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{TargetPort: intstr.IntOrString{IntVal: nexusServicePort}},
			},
		},
	}
)

func TestNewRoute(t *testing.T) {
	route, err := (&routeBuilder{}).newRoute(routeNexus, routeService).build()
	assert.Nil(t, err)
	assertRouteBasic(t, route)
}

func TestNewRouteWithRedirection(t *testing.T) {
	route, err := (&routeBuilder{}).newRoute(routeNexus, routeService).withRedirect().build()
	assert.Nil(t, err)
	assertRouteBasic(t, route)
	assertRouteRedirection(t, route)
}

func assertRouteBasic(t *testing.T, route *v1.Route) {
	assert.Equal(t, routeNexus.Name, route.Name)
	assert.Equal(t, routeNexus.Namespace, route.Namespace)

	assert.NotNil(t, route.Spec)

	assert.NotNil(t, route.Spec.To)
	assert.Equal(t, kindService, route.Spec.To.Kind)
	assert.Equal(t, routeService.Name, route.Spec.To.Name)

	assert.NotNil(t, route.Spec.Port)
	assert.Equal(t, routeService.Spec.Ports[0].TargetPort, route.Spec.Port.TargetPort)
}

func assertRouteRedirection(t *testing.T, route *v1.Route) {
	assert.NotNil(t, route.Spec.TLS)
	assert.Equal(t, v1.InsecureEdgeTerminationPolicyRedirect, route.Spec.TLS.InsecureEdgeTerminationPolicy)
	assert.Equal(t, v1.TLSTerminationEdge, route.Spec.TLS.Termination)
}
