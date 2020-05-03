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

package resource

import (
	"fmt"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	kindService  = "Service"
	routeNotInit = "route not initialized"
)

type routeBuilder struct {
	route *v1.Route
	err   error
	nexus *v1alpha1.Nexus
}

func (r *routeBuilder) newRoute(nexus *v1alpha1.Nexus, service *corev1.Service) *routeBuilder {
	route := &v1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nexus.Name,
			Namespace: nexus.Namespace,
		},
		Spec: v1.RouteSpec{
			To: v1.RouteTargetReference{
				Kind: kindService,
				Name: service.Name,
			},
			Port: &v1.RoutePort{
				TargetPort: service.Spec.Ports[0].TargetPort,
			},
		},
	}

	applyLabels(nexus, &route.ObjectMeta)
	r.route = route

	return r
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
