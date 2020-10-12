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
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/meta"
)

const (
	ingressBasePath = "/"
)

type ingressBuilder struct {
	*v1beta1.Ingress
	nexus *v1alpha1.Nexus
}

func newIngressBuilder(nexus *v1alpha1.Nexus) *ingressBuilder {
	ingress := &v1beta1.Ingress{
		ObjectMeta: meta.DefaultObjectMeta(nexus),
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: nexus.Spec.Networking.Host,
					IngressRuleValue: v1beta1.IngressRuleValue{HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{
							{
								Path: ingressBasePath,
								Backend: v1beta1.IngressBackend{
									ServiceName: nexus.Name,
									ServicePort: intstr.FromInt(deployment.NexusServicePort),
								},
							},
						},
					}},
				},
			},
		},
	}

	return &ingressBuilder{Ingress: ingress, nexus: nexus}
}

func (i *ingressBuilder) withCustomTLS() *ingressBuilder {
	i.Spec.TLS = []v1beta1.IngressTLS{
		{
			Hosts:      hosts(i.Spec.Rules),
			SecretName: i.nexus.Spec.Networking.TLS.SecretName,
		},
	}
	return i
}

func (i *ingressBuilder) build() *v1beta1.Ingress {
	return i.Ingress
}

func hosts(rules []v1beta1.IngressRule) []string {
	var hosts []string
	for _, rule := range rules {
		hosts = append(hosts, rule.Host)
	}
	return hosts
}
