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

package deployment

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	NexusServicePort = 8081
	nexusPortName    = "http"
)

func newService(nexus *v1alpha1.Nexus) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: meta.DefaultObjectMeta(nexus),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     nexusPortName,
					Protocol: corev1.ProtocolTCP,
					Port:     NexusServicePort,
					TargetPort: intstr.IntOrString{
						IntVal: NexusServicePort,
					},
				},
			},
			Selector:        meta.GenerateLabels(nexus),
			SessionAffinity: corev1.ServiceAffinityNone,
		},
	}

	if nexus.Spec.Networking.ExposeAs == v1alpha1.NodePortExposeType {
		svc.Spec.Type = corev1.ServiceTypeNodePort
		svc.Spec.Ports[0].NodePort = nexus.Spec.Networking.NodePort
		svc.Spec.ExternalTrafficPolicy = corev1.ServiceExternalTrafficPolicyTypeCluster
	}

	return svc
}
