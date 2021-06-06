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
	v1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/meta"
)

var serviceKind = (&corev1.Service{}).GroupVersionKind().Kind

type routeBuilder struct {
	*v1.Route
}

func newRouteBuilder(nexus *v1alpha1.Nexus) *routeBuilder {
	route := &v1.Route{
		ObjectMeta: meta.DefaultNetworkingMeta(nexus),
		Spec: v1.RouteSpec{
			Host: nexus.Spec.Networking.Host,
			To: v1.RouteTargetReference{
				Kind: serviceKind,
				Name: nexus.Name,
			},
			Port: &v1.RoutePort{
				TargetPort: intstr.FromString(deployment.NexusPortName),
			},
		},
	}
	return &routeBuilder{route}
}

func (r *routeBuilder) withRedirect() *routeBuilder {
	r.Spec.TLS = &v1.TLSConfig{
		Termination:                   v1.TLSTerminationEdge,
		InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
	}
	return r
}

func (r *routeBuilder) build() *v1.Route {
	return r.Route
}
