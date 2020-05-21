//     Copyright 2019 Nexus Operator and/or its authors
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
	"fmt"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	serviceKind  = "Service"
	routeNotInit = "route builder not initialized"
)

type routeBuilder struct {
	route *v1.Route
	err   error
}

func newRouteBuilder(nexus *v1alpha1.Nexus) *routeBuilder {
	route := &v1.Route{
		ObjectMeta: framework.DefaultObjectMeta(nexus),
		Spec: v1.RouteSpec{
			To: v1.RouteTargetReference{
				Kind: serviceKind,
				Name: nexus.Name,
			},
			Port: &v1.RoutePort{
				TargetPort: intstr.FromInt(deployment.NexusServicePort),
			},
		},
	}
	return &routeBuilder{route: route}
}

func (r *routeBuilder) withRedirect() *routeBuilder {
	if r == nil {
		r.err = fmt.Errorf(routeNotInit)
		return r
	}

	r.route.Spec.TLS = &v1.TLSConfig{
		Termination:                   v1.TLSTerminationEdge,
		InsecureEdgeTerminationPolicy: v1.InsecureEdgeTerminationPolicyRedirect,
	}
	return r
}

func (r *routeBuilder) build() (*v1.Route, error) {
	return r.route, r.err
}
