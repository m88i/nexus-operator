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

package networking

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/deployment"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"testing"
)

var (
	ingressNexus = &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nexus3",
			Namespace: "nexus",
		},
		Spec: v1alpha1.NexusSpec{
			Networking: v1alpha1.NexusNetworking{
				Host: "ingress.tls.test.com",
				TLS: v1alpha1.NexusNetworkingTLS{
					SecretName: "test-tls",
				},
			},
		},
	}

	ingressService = &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nexus3",
			Namespace: "nexus",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{TargetPort: intstr.IntOrString{IntVal: deployment.NexusServicePort}},
			},
		},
	}
)

func TestHosts(t *testing.T) {
	ingress := &v1beta1.Ingress{
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{},
		},
	}

	h := hosts(ingress)
	assert.Len(t, h, 0)

	host := "a"
	ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{Host: host})
	h = hosts(ingress)
	assert.Len(t, h, 1)
	assert.Equal(t, h[0], host)

	host = "b"
	ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{Host: host})
	h = hosts(ingress)
	assert.Len(t, h, 2)
	assert.Equal(t, h[1], host)
}

func TestNewIngress(t *testing.T) {
	ingress, err := newIngressBuilder(ingressNexus).build()
	assert.Nil(t, err)
	assertIngressBasic(t, ingress)
}

func TestNewIngressWithSecretName(t *testing.T) {
	ingress, err := newIngressBuilder(ingressNexus).withCustomTLS().build()
	assert.Nil(t, err)
	assertIngressBasic(t, ingress)
	assertIngressSecretName(t, ingress)
}

func assertIngressBasic(t *testing.T, ingress *v1beta1.Ingress) {
	assert.Equal(t, ingressNexus.Name, ingress.Name)
	assert.Equal(t, ingressNexus.Namespace, ingress.Namespace)

	assert.NotNil(t, ingress.Spec)

	assert.Len(t, ingress.Spec.Rules, 1)
	rule := ingress.Spec.Rules[0]

	assert.Equal(t, ingressNexus.Spec.Networking.Host, rule.Host)
	assert.NotNil(t, rule.IngressRuleValue)
	assert.NotNil(t, rule.IngressRuleValue.HTTP)

	assert.Len(t, rule.IngressRuleValue.HTTP.Paths, 1)
	path := rule.IngressRuleValue.HTTP.Paths[0]

	assert.Equal(t, ingressBasePath, path.Path)
	assert.NotNil(t, path.Backend)
	assert.Equal(t, ingressService.Spec.Ports[0].TargetPort, path.Backend.ServicePort)
	assert.Equal(t, ingressService.Name, path.Backend.ServiceName)
}

func assertIngressSecretName(t *testing.T, ingress *v1beta1.Ingress) {
	assert.Len(t, ingress.Spec.TLS, 1)
	assert.Equal(t, ingressNexus.Spec.Networking.TLS.SecretName, ingress.Spec.TLS[0].SecretName)

	assert.Len(t, ingress.Spec.TLS[0].Hosts, 1)
	assert.Equal(t, ingressNexus.Spec.Networking.Host, ingress.Spec.TLS[0].Hosts[0])
}
