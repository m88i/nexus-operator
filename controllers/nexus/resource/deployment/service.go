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

package deployment

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/meta"
)

const (
	// NexusPortName is the name of the port on the service
	NexusPortName = "http"
	// DefaultHTTPPort is the default HTTP port
	DefaultHTTPPort    = 80
	nexusContainerPort = 8081
)

func newService(nexus *v1alpha1.Nexus) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: meta.DefaultObjectMeta(nexus),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     NexusPortName,
					Protocol: corev1.ProtocolTCP,
					Port:     DefaultHTTPPort,
					TargetPort: intstr.IntOrString{
						IntVal: nexusContainerPort,
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
