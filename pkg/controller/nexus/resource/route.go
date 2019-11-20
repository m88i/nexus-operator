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
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	v1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const kindService = "Service"

func newRoute(nexus *v1alpha1.Nexus, service *corev1.Service) *v1.Route {
	return &v1.Route{
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
}
