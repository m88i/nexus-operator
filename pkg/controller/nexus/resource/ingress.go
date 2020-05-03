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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	ingressBasePath = "/"
	ingressNotInit  = "ingress not initialized"
)

type ingressBuilder struct {
	ingress *v1beta1.Ingress
	err     error
	nexus   *v1alpha1.Nexus
}

func (i *ingressBuilder) newIngress(nexus *v1alpha1.Nexus, service *corev1.Service) *ingressBuilder {
	port, err := getNexusDefaultPort(service)
	if err != nil {
		i.err = err
		return i
	}

	ingress := &v1beta1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nexus.Name,
			Namespace: nexus.Namespace,
		},
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{
				{
					Host: nexus.Spec.Networking.Host,
					IngressRuleValue: v1beta1.IngressRuleValue{HTTP: &v1beta1.HTTPIngressRuleValue{
						Paths: []v1beta1.HTTPIngressPath{
							{
								Path: ingressBasePath,
								Backend: v1beta1.IngressBackend{
									ServiceName: service.Name,
									ServicePort: port,
								},
							},
						},
					}},
				},
			},
		},
	}

	applyLabels(nexus, &ingress.ObjectMeta)
	i.ingress = ingress
	i.nexus = nexus

	return i
}

func (i *ingressBuilder) withCustomTLS() *ingressBuilder {
	if i == nil {
		i.err = fmt.Errorf(ingressNotInit)
		return i
	}

	i.ingress.Spec.TLS = []v1beta1.IngressTLS{
		{
			Hosts:      hosts(i.ingress),
			SecretName: i.nexus.Spec.Networking.TLS.SecretName,
		},
	}
	return i
}

func (i *ingressBuilder) build() (*v1beta1.Ingress, error) {
	if i == nil {
		return nil, fmt.Errorf(ingressNotInit)
	}

	if i.err != nil {
		return nil, i.err
	}

	return i.ingress, nil
}

func getNexusDefaultPort(service *corev1.Service) (intstr.IntOrString, error) {
	for _, port := range service.Spec.Ports {
		if port.TargetPort.IntVal == nexusServicePort {
			return port.TargetPort, nil
		}
	}
	return intstr.IntOrString{IntVal: 0}, fmt.Errorf("No default nexus port (%d) found in service %s ", nexusServicePort, service.Name)
}

func hosts(ingress *v1beta1.Ingress) []string {
	var hosts []string
	for _, rule := range ingress.Spec.Rules {
		hosts = append(hosts, rule.Host)
	}
	return hosts
}
