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
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/meta"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
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
