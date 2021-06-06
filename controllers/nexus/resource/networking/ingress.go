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
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/deployment"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/meta"
)

const (
	ingressBasePath   = "/?(.*)"
	ingressClassKey   = "kubernetes.io/ingress.class"
	ingressClassNginx = "nginx"
	nginxRewriteKey   = "nginx.ingress.kubernetes.io/rewrite-target"
	nginxRewriteValue = "/$1"
)

// hack to take the address of v1.PathExactType
var pathTypePrefix = v1.PathTypePrefix

type ingressBuilder struct {
	*v1.Ingress
	nexus *v1alpha1.Nexus
}

func newIngressBuilder(nexus *v1alpha1.Nexus) *ingressBuilder {
	ingress := &v1.Ingress{
		ObjectMeta: meta.DefaultNetworkingMeta(nexus),
		Spec: v1.IngressSpec{
			Rules: []v1.IngressRule{
				{
					Host: nexus.Spec.Networking.Host,
					IngressRuleValue: v1.IngressRuleValue{
						HTTP: &v1.HTTPIngressRuleValue{
							Paths: []v1.HTTPIngressPath{
								{
									PathType: &pathTypePrefix,
									Path:     ingressBasePath,
									Backend: v1.IngressBackend{
										Service: &v1.IngressServiceBackend{
											Name: nexus.Name,
											Port: v1.ServiceBackendPort{Number: deployment.DefaultHTTPPort},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	addNginxAnnotations(ingress.ObjectMeta)
	return &ingressBuilder{Ingress: ingress, nexus: nexus}
}

func (i *ingressBuilder) withCustomTLS() *ingressBuilder {
	i.Spec.TLS = []v1.IngressTLS{
		{
			Hosts:      hosts(i.Spec.Rules),
			SecretName: i.nexus.Spec.Networking.TLS.SecretName,
		},
	}
	return i
}

func (i *ingressBuilder) build() *v1.Ingress {
	return i.Ingress
}

func hosts(rules []v1.IngressRule) []string {
	var hosts []string
	for _, rule := range rules {
		hosts = append(hosts, rule.Host)
	}
	return hosts
}

func addNginxAnnotations(meta metav1.ObjectMeta) {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}

	meta.Annotations[ingressClassKey] = ingressClassNginx
	meta.Annotations[nginxRewriteKey] = nginxRewriteValue
}
