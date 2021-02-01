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

package v1alpha1

import (
	"fmt"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"

	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name   string
		client *discovery.FakeDisc
		want   *validator
	}{
		{
			"On OCP",
			discovery.NewFakeDiscBuilder().OnOpenshift().Build(),
			&validator{
				routeAvailable:   true,
				ingressAvailable: false,
			},
		},
		{
			"On K8s, no ingress",
			discovery.NewFakeDiscBuilder().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: false,
			},
		},
		{
			"On K8s with v1beta1 ingress",
			discovery.NewFakeDiscBuilder().WithLegacyIngress().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: true,
			},
		},
		{
			"On K8s with v1 ingress",
			discovery.NewFakeDiscBuilder().WithIngress().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: true,
			},
		},
	}

	// the nexus itself isn't important, but we need it to avoid nil dereference
	n := &Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: t.Name()}}
	for _, tt := range tests {
		discovery.SetClient(tt.client)
		v := newValidator(n)
		assert.Equal(t, tt.want.ingressAvailable, v.ingressAvailable)
		assert.Equal(t, tt.want.routeAvailable, v.routeAvailable)
	}

	// now let's test out error
	client := discovery.NewFakeDiscBuilder().WithIngress().Build()
	errString := "test error"
	client.SetMockError(fmt.Errorf(errString))
	discovery.SetClient(client)
	got := newValidator(n)
	// all calls to discovery failed, so all should be false
	assert.False(t, got.routeAvailable)
	assert.False(t, got.ingressAvailable)
}

func TestValidator_validate_Networking(t *testing.T) {
	tests := []struct {
		name    string
		disc    *discovery.FakeDisc
		input   *Nexus
		wantErr bool
	}{
		{
			"'spec.networking.expose' left blank or set to false",
			discovery.NewFakeDiscBuilder().Build(),

			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: false}}},
			false,
		},
		{
			"Valid Nexus with Ingress on K8s",
			discovery.NewFakeDiscBuilder().WithIngress().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com"}}},
			false,
		},
		{
			"Valid Nexus with Ingress and TLS secret on K8s",
			discovery.NewFakeDiscBuilder().WithIngress().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com", TLS: NexusNetworkingTLS{SecretName: "tt-secret"}}}},
			false,
		},
		{
			"Valid Nexus with Ingress on K8s, but Ingress unavailable (Kubernetes < 1.14)",
			discovery.NewFakeDiscBuilder().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com"}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on K8s and no 'spec.networking.host'",
			discovery.NewFakeDiscBuilder().WithIngress().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on K8s and 'spec.networking.mandatory' set to 'true'",
			discovery.NewFakeDiscBuilder().WithIngress().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com", TLS: NexusNetworkingTLS{Mandatory: true}}}},
			true,
		},
		{
			"Invalid Nexus with Route on K8s",
			discovery.NewFakeDiscBuilder().WithIngress().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: RouteExposeType}}},
			true,
		},
		{
			"Valid Nexus with Route on OCP",
			discovery.NewFakeDiscBuilder().OnOpenshift().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: RouteExposeType}}},
			false,
		},
		{
			"Valid Nexus with Route on OCP with mandatory TLS",
			discovery.NewFakeDiscBuilder().OnOpenshift().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: RouteExposeType, TLS: NexusNetworkingTLS{Mandatory: true}}}},
			false,
		},
		{
			"Invalid Nexus with Route on OCP, but using secret name",
			discovery.NewFakeDiscBuilder().OnOpenshift().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: RouteExposeType, TLS: NexusNetworkingTLS{SecretName: "test-secret"}}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on OCP",
			discovery.NewFakeDiscBuilder().OnOpenshift().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com"}}},
			true,
		},
		{
			"Valid Nexus with Node Port",
			discovery.NewFakeDiscBuilder().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: NodePortExposeType, NodePort: 8080}}},
			false,
		},
		{
			"Invalid Nexus with Node Port and no port informed",
			discovery.NewFakeDiscBuilder().Build(),
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: NodePortExposeType}}},
			true,
		},
	}

	for _, tt := range tests {
		discovery.SetClient(tt.disc)
		err := newValidator(tt.input).validateNetworking()
		if (err != nil) != tt.wantErr {
			t.Errorf("%s\nnexus: %#v\nwantErr: %v\terr: %#v\n", tt.name, tt.input, tt.wantErr, err)
		}
	}
}

func TestValidator_validate_Persistence(t *testing.T) {
	tests := []struct {
		name    string
		input   *Nexus
		wantErr bool
	}{
		{
			"Invalid volume size",
			&Nexus{Spec: NexusSpec{Persistence: NexusPersistence{VolumeSize: "not-a-quantity"}}},
			true,
		},
		{
			"Valid volume size",
			&Nexus{Spec: NexusSpec{Persistence: NexusPersistence{VolumeSize: "10Gi"}}},
			false,
		},
	}

	discovery.SetClient(discovery.NewFakeDiscBuilder().Build())
	for _, tt := range tests {
		err := newValidator(tt.input).validatePersistence()
		if (err != nil) != tt.wantErr {
			t.Errorf("%s\nnexus: %#v\nwantErr: %v\terr: %#v\n", tt.name, tt.input, tt.wantErr, err)
		}
	}
}
