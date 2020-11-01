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
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/deployment"
)

var (
	nexusIngress = &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nexus3",
			Namespace: "nexusIngress",
		},
		Spec: v1alpha1.NexusSpec{
			Networking: v1alpha1.NexusNetworking{
				Expose:   true,
				ExposeAs: v1alpha1.IngressExposeType,
				Host:     "ingress.tls.test.com",
				TLS: v1alpha1.NexusNetworkingTLS{
					SecretName: "test-tls",
				},
			},
		},
	}
)

func TestHosts(t *testing.T) {
	ingress := &v1.Ingress{
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{},
		},
	}

	h := hosts(ingress.Spec.Rules)
	assert.Len(t, h, 0)

	host := "a"
	ingress.Spec.Rules = append(ingress.Spec.Rules, v1.IngressRule{Host: host})
	h = hosts(ingress.Spec.Rules)
	assert.Len(t, h, 1)
	assert.Equal(t, h[0], host)

	host = "b"
	ingress.Spec.Rules = append(ingress.Spec.Rules, v1.IngressRule{Host: host})
	h = hosts(ingress.Spec.Rules)
	assert.Len(t, h, 2)
	assert.Equal(t, h[1], host)
}

func TestNewIngress(t *testing.T) {
	ingress := newIngressBuilder(nexusIngress).build()
	assertIngressBasic(t, ingress)
}

func TestNewIngressWithSecretName(t *testing.T) {
	ingress := newIngressBuilder(nexusIngress).withCustomTLS().build()
	assertIngressBasic(t, ingress)
	assertIngressSecretName(t, ingress)
}

func assertIngressBasic(t *testing.T, ingress *v1.Ingress) {
	assert.Equal(t, nexusIngress.Name, ingress.Name)
	assert.Equal(t, nexusIngress.Namespace, ingress.Namespace)

	assert.NotNil(t, ingress.Spec)

	assert.Len(t, ingress.Spec.Rules, 1)
	rule := ingress.Spec.Rules[0]

	assert.Equal(t, nexusIngress.Spec.Networking.Host, rule.Host)
	assert.NotNil(t, rule.IngressRuleValue)
	assert.NotNil(t, rule.IngressRuleValue.HTTP)

	assert.Len(t, rule.IngressRuleValue.HTTP.Paths, 1)
	path := rule.IngressRuleValue.HTTP.Paths[0]

	assert.Equal(t, ingressBasePath, path.Path)
	assert.NotNil(t, path.Backend)
	assert.Equal(t, int32(deployment.DefaultHTTPPort), path.Backend.Service.Port.Number)
	assert.Equal(t, nexusIngress.Name, path.Backend.Service.Name)
	assert.NotEmpty(t, ingress.Annotations[nginxRewriteKey])
	assert.Equal(t, ingressClassNginx, ingress.Annotations[ingressClassKey])
}

func assertIngressSecretName(t *testing.T, ingress *v1.Ingress) {
	assert.Len(t, ingress.Spec.TLS, 1)
	assert.Equal(t, nexusIngress.Spec.Networking.TLS.SecretName, ingress.Spec.TLS[0].SecretName)

	assert.Len(t, ingress.Spec.TLS[0].Hosts, 1)
	assert.Equal(t, nexusIngress.Spec.Networking.Host, ingress.Spec.TLS[0].Hosts[0])
}
