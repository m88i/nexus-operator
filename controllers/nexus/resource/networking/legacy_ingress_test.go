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
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/m88i/nexus-operator/controllers/nexus/resource/deployment"
)

func TestLegacyHosts(t *testing.T) {
	ingress := &v1beta1.Ingress{
		Spec: v1beta1.IngressSpec{
			Rules: []v1beta1.IngressRule{},
		},
	}

	h := legacyHosts(ingress.Spec.Rules)
	assert.Len(t, h, 0)

	host := "a"
	ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{Host: host})
	h = legacyHosts(ingress.Spec.Rules)
	assert.Len(t, h, 1)
	assert.Equal(t, h[0], host)

	host = "b"
	ingress.Spec.Rules = append(ingress.Spec.Rules, v1beta1.IngressRule{Host: host})
	h = legacyHosts(ingress.Spec.Rules)
	assert.Len(t, h, 2)
	assert.Equal(t, h[1], host)
}

func TestNewLegacyIngress(t *testing.T) {
	ingress := newLegacyIngressBuilder(nexusIngress).build()
	assertLegacyIngressBasic(t, ingress)
}

func TestNewLegacyIngressWithSecretName(t *testing.T) {
	ingress := newLegacyIngressBuilder(nexusIngress).withCustomTLS().build()
	assertLegacyIngressBasic(t, ingress)
	assertLegacyIngressSecretName(t, ingress)
}

func assertLegacyIngressBasic(t *testing.T, ingress *v1beta1.Ingress) {
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
	assert.Equal(t, intstr.FromInt(deployment.DefaultHTTPPort), path.Backend.ServicePort)
	assert.Equal(t, nexusIngress.Name, path.Backend.ServiceName)
	assert.NotEmpty(t, ingress.Annotations[nginxRewriteKey])
	assert.Equal(t, ingressClassNginx, ingress.Annotations[ingressClassKey])
}

func assertLegacyIngressSecretName(t *testing.T, ingress *v1beta1.Ingress) {
	assert.Len(t, ingress.Spec.TLS, 1)
	assert.Equal(t, nexusIngress.Spec.Networking.TLS.SecretName, ingress.Spec.TLS[0].SecretName)

	assert.Len(t, ingress.Spec.TLS[0].Hosts, 1)
	assert.Equal(t, nexusIngress.Spec.Networking.Host, ingress.Spec.TLS[0].Hosts[0])
}
