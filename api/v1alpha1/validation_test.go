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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/m88i/nexus-operator/pkg/logger"
)

// TODO: fix the tests
// TODO: move mutation tests to mutation_test.go

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name   string
		client *test.FakeClient
		want   *validator
	}{
		{
			"On OCP",
			test.NewFakeClientBuilder().OnOpenshift().Build(),
			&validator{
				routeAvailable:   true,
				ingressAvailable: false,
			},
		},
		{
			"On K8s, no ingress",
			test.NewFakeClientBuilder().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: false,
			},
		},
		{
			"On K8s with v1beta1 ingress",
			test.NewFakeClientBuilder().WithLegacyIngress().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: true,
			},
		},
		{
			"On K8s with v1 ingress",
			test.NewFakeClientBuilder().WithIngress().Build(),
			&validator{
				routeAvailable:   false,
				ingressAvailable: true,
			},
		},
	}

	for _, tt := range tests {
		discovery.SetClient(tt.client)
		// the nexus CR is not important here
		assert.Equal(t, tt.want, newValidator(nil))
	}

	// now let's test out error
	client := test.NewFakeClientBuilder().WithIngress().Build()
	errString := "test error"
	client.SetMockError(fmt.Errorf(errString))
	discovery.SetClient(client)
	got := newValidator(nil)
	// all calls to discovery failed, so all should be false
	assert.False(t, got.routeAvailable)
	assert.False(t, got.ingressAvailable)
}

func TestValidator_SetDefaultsAndValidate_Deployment(t *testing.T) {
	minimumDefaultProbe := &NexusProbe{
		InitialDelaySeconds: 0,
		TimeoutSeconds:      1,
		PeriodSeconds:       1,
		SuccessThreshold:    1,
		FailureThreshold:    1,
	}

	tests := []struct {
		name  string
		input *Nexus
		want  *Nexus
	}{
		{
			"'spec.resources' left blank",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Resources = corev1.ResourceRequirements{}
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.useRedHatImage' set to true and 'spec.image' not left blank",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.UseRedHatImage = true
				nexus.Spec.Image = "some-image"
				return nexus
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.UseRedHatImage = true
				n.Spec.Image = NexusCertifiedImage
				return n
			}(),
		},
		{
			"'spec.useRedHatImage' set to false and 'spec.image' left blank",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Image = ""
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.successThreshold' not equal to 1",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe.SuccessThreshold = 2
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.*' and 'spec.readinessProbe.*' don't meet minimum values",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = &NexusProbe{
					InitialDelaySeconds: -1,
					TimeoutSeconds:      -1,
					PeriodSeconds:       -1,
					SuccessThreshold:    -1,
					FailureThreshold:    -1,
				}
				nexus.Spec.ReadinessProbe = nexus.Spec.LivenessProbe.DeepCopy()
				return nexus
			}(),
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = minimumDefaultProbe.DeepCopy()
				nexus.Spec.ReadinessProbe = minimumDefaultProbe.DeepCopy()
				return nexus
			}(),
		},
		{
			"Unset 'spec.livenessProbe' and 'spec.readinessProbe'",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = nil
				nexus.Spec.ReadinessProbe = nil
				return nexus
			}(),
			AllDefaultsCommunityNexus.DeepCopy(),
		},
		{
			"Invalid 'spec.imagePullPolicy'",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ImagePullPolicy = "invalid"
				return nexus
			}(),
			AllDefaultsCommunityNexus.DeepCopy(),
		},
		{
			"Invalid 'spec.replicas'",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Replicas = 3
				return nexus
			}(),
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Replicas = 1
				return nexus
			}(),
		},
	}

	for _, tt := range tests {
		v := &validator{}
		got, err := v.SetDefaultsAndValidate(tt.input)
		assert.Nil(t, err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s\nWant: %+v\nGot: %+v", tt.name, tt.want, got)
		}
	}
}

func TestValidator_setUpdateDefaults(t *testing.T) {
	client := test.NewFakeClientBuilder().Build()
	nexus := &Nexus{Spec: NexusSpec{AutomaticUpdate: NexusAutomaticUpdate{}}}
	nexus.Spec.Image = NexusCommunityImage
	v, _ := newValidator(client, client.Scheme())
	v.log = logger.GetLoggerWithResource("test", nexus)

	v.setUpdateDefaults(nexus)
	latestMinor, err := framework.GetLatestMinor()
	if err != nil {
		// If we couldn't fetch the tags updates should be disabled
		assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)
	} else {
		assert.Equal(t, latestMinor, *nexus.Spec.AutomaticUpdate.MinorVersion)
	}

	// Now an invalid image
	nexus = &Nexus{Spec: NexusSpec{AutomaticUpdate: NexusAutomaticUpdate{}}}
	nexus.Spec.Image = "some-image"
	v.setUpdateDefaults(nexus)
	assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)

	// Informed a minor which does not exist
	nexus = &Nexus{Spec: NexusSpec{AutomaticUpdate: NexusAutomaticUpdate{}}}
	nexus.Spec.Image = NexusCommunityImage
	bogusMinor := -1
	nexus.Spec.AutomaticUpdate.MinorVersion = &bogusMinor
	v.setUpdateDefaults(nexus)
	latestMinor, err = framework.GetLatestMinor()
	if err != nil {
		// If we couldn't fetch the tags updates should be disabled
		assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)
	} else {
		assert.Equal(t, latestMinor, *nexus.Spec.AutomaticUpdate.MinorVersion)
	}
}

func TestValidator_setNetworkingDefaults(t *testing.T) {
	tests := []struct {
		name             string
		ocp              bool
		routeAvailable   bool
		ingressAvailable bool
		input            *Nexus
		want             *Nexus
	}{
		{
			"'spec.networking.exposeAs' left blank on OCP",
			true,
			true,
			false, // unimportant
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = RouteExposeType
				return n
			}(),
		},
		{
			"'spec.networking.exposeAs' left blank on K8s",
			false,
			false,
			true,
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = IngressExposeType
				return n
			}(),
		},
		{
			"'spec.networking.exposeAs' left blank on K8s, but Ingress unavailable",
			false,
			false,
			false,
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = NodePortExposeType
				return n
			}(),
		},
	}

	for _, tt := range tests {
		v := &validator{
			routeAvailable:   tt.routeAvailable,
			ingressAvailable: tt.ingressAvailable,
			ocp:              tt.ocp,
			log:              logger.GetLoggerWithResource("test", tt.input),
		}
		got := tt.input.DeepCopy()
		v.setNetworkingDefaults(got)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s\nWant: %v\nGot: %v", tt.name, tt.want, got)
		}
	}
}

func TestValidator_validateNetworking(t *testing.T) {
	tests := []struct {
		name             string
		ocp              bool
		routeAvailable   bool
		ingressAvailable bool
		input            *Nexus
		wantError        bool
	}{
		{
			"'spec.networking.expose' left blank or set to false",
			false, // unimportant
			false, // unimportant
			false, // unimportant
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: false}}},
			false,
		},
		{
			"Valid Nexus with Ingress on K8s",
			false,
			false,
			true,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com"}}},
			false,
		},
		{
			"Valid Nexus with Ingress and TLS secret on K8s",
			false,
			false,
			true,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com", TLS: NexusNetworkingTLS{SecretName: "tt-secret"}}}},
			false,
		},
		{
			"Valid Nexus with Ingress on K8s, but Ingress unavailable (Kubernetes < 1.14)",
			false,
			false,
			false,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com"}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on K8s and no 'spec.networking.host'",
			false,
			false,
			true,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on K8s and 'spec.networking.mandatory' set to 'true'",
			false,
			false,
			true,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com", TLS: NexusNetworkingTLS{Mandatory: true}}}},
			true,
		},
		{
			"Invalid Nexus with Route on K8s",
			false,
			false,
			true,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: RouteExposeType}}},
			true,
		},
		{
			"Valid Nexus with Route on OCP",
			true,
			true,
			false,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: RouteExposeType}}},
			false,
		},
		{
			"Valid Nexus with Route on OCP with mandatory TLS",
			true,
			true,
			false,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: RouteExposeType, TLS: NexusNetworkingTLS{Mandatory: true}}}},
			false,
		},
		{
			"Invalid Nexus with Route on OCP, but using secret name",
			true,
			true,
			false,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: RouteExposeType, TLS: NexusNetworkingTLS{SecretName: "test-secret"}}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on OCP",
			true,
			true,
			false,
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: IngressExposeType, Host: "example.com"}}},
			true,
		},
		{
			"Valid Nexus with Node Port",
			false, // unimportant
			false, // unimportant
			false, // unimportant
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: NodePortExposeType, NodePort: 8080}}},
			false,
		},
		{
			"Invalid Nexus with Node Port and no port informed",
			false, // unimportant
			false, // unimportant
			false, // unimportant
			&Nexus{Spec: NexusSpec{Networking: NexusNetworking{Expose: true, ExposeAs: NodePortExposeType}}},
			true,
		},
	}

	for _, tt := range tests {
		v := &validator{
			routeAvailable:   tt.routeAvailable,
			ingressAvailable: tt.ingressAvailable,
			ocp:              tt.ocp,
			log:              logger.GetLoggerWithResource("test", tt.input),
		}
		if err := v.validateNetworking(tt.input); (err != nil) != tt.wantError {
			t.Errorf("%s\nWantError: %v\tError: %v", tt.name, tt.wantError, err)
		}
	}
}

func TestValidator_SetDefaultsAndValidate_Persistence(t *testing.T) {
	tests := []struct {
		name  string
		input *Nexus
		want  *Nexus
	}{
		{
			"'spec.persistence.volumeSize' left blank",
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Persistence.Persistent = true
				return n
			}(),
			func() *Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Persistence.Persistent = true
				n.Spec.Persistence.VolumeSize = DefaultVolumeSize
				return n
			}(),
		},
	}
	for _, tt := range tests {
		v := &validator{}
		got, err := v.SetDefaultsAndValidate(tt.input)
		assert.Nil(t, err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s\nWant: %+v\nGot: %+v", tt.name, tt.want, got)
		}
	}
}

func TestValidator_SetDefaultsAndValidate_Security(t *testing.T) {

	tests := []struct {
		name  string
		input *Nexus
		want  *Nexus
	}{
		{
			"'spec.serviceAccountName' left blank",
			func() *Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ServiceAccountName = ""
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
	}
	for _, tt := range tests {
		v := &validator{}
		got, err := v.SetDefaultsAndValidate(tt.input)
		assert.Nil(t, err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s\nWant: %+v\nGot: %+v", tt.name, tt.want, got)
		}
	}
}
