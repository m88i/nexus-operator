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

package validation

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/update"
	"github.com/m88i/nexus-operator/pkg/logger"
	"github.com/m88i/nexus-operator/pkg/test"
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name   string
		client *test.FakeClient
		want   *Validator
	}{
		{
			"On OCP",
			test.NewFakeClientBuilder().OnOpenshift().Build(),
			&Validator{
				routeAvailable:   true,
				ingressAvailable: false,
				ocp:              true,
			},
		},
		{
			"On K8s, no ingress",
			test.NewFakeClientBuilder().Build(),
			&Validator{
				routeAvailable:   false,
				ingressAvailable: false,
				ocp:              false,
			},
		},
		{
			"On K8s with ingress",
			test.NewFakeClientBuilder().WithIngress().Build(),
			&Validator{
				routeAvailable:   false,
				ingressAvailable: true,
				ocp:              false,
			},
		},
	}

	for _, tt := range tests {
		got, err := NewValidator(tt.client, tt.client.Scheme(), tt.client)
		assert.Nil(t, err)
		tt.want.client = tt.client
		tt.want.scheme = tt.client.Scheme()
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s\nWant: %+v\nGot: %+v", tt.name, tt.want, got)
		}
	}

	// now let's test out error
	client := test.NewFakeClientBuilder().Build()
	errString := "test error"
	client.SetMockErrorForOneRequest(fmt.Errorf(errString))
	_, err := NewValidator(client, client.Scheme(), client)
	assert.Contains(t, err.Error(), errString)
}

func TestValidator_SetDefaultsAndValidate_Deployment(t *testing.T) {
	minimumDefaultProbe := &v1alpha1.NexusProbe{
		InitialDelaySeconds: 0,
		TimeoutSeconds:      1,
		PeriodSeconds:       1,
		SuccessThreshold:    1,
		FailureThreshold:    1,
	}

	tests := []struct {
		name  string
		input *v1alpha1.Nexus
		want  *v1alpha1.Nexus
	}{
		{
			"'spec.resources' left blank",
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Resources = corev1.ResourceRequirements{}
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.useRedHatImage' set to true and 'spec.image' not left blank",
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.UseRedHatImage = true
				nexus.Spec.Image = "some-image"
				return nexus
			}(),
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.UseRedHatImage = true
				n.Spec.Image = NexusCertifiedImage
				return n
			}(),
		},
		{
			"'spec.useRedHatImage' set to false and 'spec.image' left blank",
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Image = ""
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.successThreshold' not equal to 1",
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe.SuccessThreshold = 2
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.*' and 'spec.readinessProbe.*' don't meet minimum values",
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = &v1alpha1.NexusProbe{
					InitialDelaySeconds: -1,
					TimeoutSeconds:      -1,
					PeriodSeconds:       -1,
					SuccessThreshold:    -1,
					FailureThreshold:    -1,
				}
				nexus.Spec.ReadinessProbe = nexus.Spec.LivenessProbe.DeepCopy()
				return nexus
			}(),
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = minimumDefaultProbe.DeepCopy()
				nexus.Spec.ReadinessProbe = minimumDefaultProbe.DeepCopy()
				return nexus
			}(),
		},
		{
			"Unset 'spec.livenessProbe' and 'spec.readinessProbe'",
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = nil
				nexus.Spec.ReadinessProbe = nil
				return nexus
			}(),
			AllDefaultsCommunityNexus.DeepCopy(),
		},
		{
			"Invalid 'spec.imagePullPolicy'",
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ImagePullPolicy = "invalid"
				return nexus
			}(),
			AllDefaultsCommunityNexus.DeepCopy(),
		},
	}

	for _, tt := range tests {
		v := &Validator{}
		got, err := v.SetDefaultsAndValidate(tt.input)
		assert.Nil(t, err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s\nWant: %+v\nGot: %+v", tt.name, tt.want, got)
		}
	}
}

func TestValidator_setUpdateDefaults(t *testing.T) {
	client := test.NewFakeClientBuilder().Build()
	nexus := &v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{AutomaticUpdate: v1alpha1.NexusAutomaticUpdate{}}}
	nexus.Spec.Image = NexusCommunityImage
	v, _ := NewValidator(client, client.Scheme(), client)
	v.log = logger.GetLoggerWithResource("test", nexus)

	v.setUpdateDefaults(nexus)
	latestMinor, err := update.GetLatestMinor()
	if err != nil {
		// If we couldn't fetch the tags updates should be disabled
		assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)
	} else {
		assert.Equal(t, latestMinor, *nexus.Spec.AutomaticUpdate.MinorVersion)
	}

	// Now an invalid image
	nexus = &v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{AutomaticUpdate: v1alpha1.NexusAutomaticUpdate{}}}
	nexus.Spec.Image = "some-image"
	v.setUpdateDefaults(nexus)
	assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)

	// Informed a minor which does not exist
	nexus = &v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{AutomaticUpdate: v1alpha1.NexusAutomaticUpdate{}}}
	nexus.Spec.Image = NexusCommunityImage
	bogusMinor := -1
	nexus.Spec.AutomaticUpdate.MinorVersion = &bogusMinor
	v.setUpdateDefaults(nexus)
	latestMinor, err = update.GetLatestMinor()
	if err != nil {
		// If we couldn't fetch the tags updates should be disabled
		assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)
		assert.True(t, test.EventExists(client, changedNexusReason))
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
		input            *v1alpha1.Nexus
		want             *v1alpha1.Nexus
	}{
		{
			"'spec.networking.exposeAs' left blank on OCP",
			true,
			true,
			false, // unimportant
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = v1alpha1.RouteExposeType
				return n
			}(),
		},
		{
			"'spec.networking.exposeAs' left blank on K8s",
			false,
			false,
			true,
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = v1alpha1.IngressExposeType
				return n
			}(),
		},
		{
			"'spec.networking.exposeAs' left blank on K8s, but Ingress unavailable",
			false,
			false,
			false,
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				n.Spec.Networking.ExposeAs = v1alpha1.NodePortExposeType
				return n
			}(),
		},
	}

	for _, tt := range tests {
		v := &Validator{
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
		input            *v1alpha1.Nexus
		wantError        bool
	}{
		{
			"'spec.networking.expose' left blank or set to false",
			false, // unimportant
			false, // unimportant
			false, // unimportant
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: false}}},
			false,
		},
		{
			"Valid Nexus with Ingress on K8s",
			false,
			false,
			true,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com"}}},
			false,
		},
		{
			"Valid Nexus with Ingress and TLS secret on K8s",
			false,
			false,
			true,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com", TLS: v1alpha1.NexusNetworkingTLS{SecretName: "tt-secret"}}}},
			false,
		},
		{
			"Valid Nexus with Ingress on K8s, but Ingress unavailable (Kubernetes < 1.14)",
			false,
			false,
			false,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com"}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on K8s and no 'spec.networking.host'",
			false,
			false,
			true,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on K8s and 'spec.networking.mandatory' set to 'true'",
			false,
			false,
			true,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com", TLS: v1alpha1.NexusNetworkingTLS{Mandatory: true}}}},
			true,
		},
		{
			"Invalid Nexus with Route on K8s",
			false,
			false,
			true,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.RouteExposeType}}},
			true,
		},
		{
			"Valid Nexus with Route on OCP",
			true,
			true,
			false,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.RouteExposeType}}},
			false,
		},
		{
			"Valid Nexus with Route on OCP with mandatory TLS",
			true,
			true,
			false,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.RouteExposeType, TLS: v1alpha1.NexusNetworkingTLS{Mandatory: true}}}},
			false,
		},
		{
			"Invalid Nexus with Route on OCP, but using secret name",
			true,
			true,
			false,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.RouteExposeType, TLS: v1alpha1.NexusNetworkingTLS{SecretName: "test-secret"}}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on OCP",
			true,
			true,
			false,
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com"}}},
			true,
		},
		{
			"Valid Nexus with Node Port",
			false, // unimportant
			false, // unimportant
			false, // unimportant
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.NodePortExposeType, NodePort: 8080}}},
			false,
		},
		{
			"Invalid Nexus with Node Port and no port informed",
			false, // unimportant
			false, // unimportant
			false, // unimportant
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.NodePortExposeType}}},
			true,
		},
	}

	for _, tt := range tests {
		v := &Validator{
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
		input *v1alpha1.Nexus
		want  *v1alpha1.Nexus
	}{
		{
			"'spec.persistence.volumeSize' left blank",
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Persistence.Persistent = true
				return n
			}(),
			func() *v1alpha1.Nexus {
				n := AllDefaultsCommunityNexus.DeepCopy()
				n.Spec.Persistence.Persistent = true
				n.Spec.Persistence.VolumeSize = DefaultVolumeSize
				return n
			}(),
		},
	}
	for _, tt := range tests {
		v := &Validator{}
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
		input *v1alpha1.Nexus
		want  *v1alpha1.Nexus
	}{
		{
			"'spec.serviceAccountName' left blank",
			func() *v1alpha1.Nexus {
				nexus := AllDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ServiceAccountName = ""
				return nexus
			}(),
			&AllDefaultsCommunityNexus,
		},
	}
	for _, tt := range tests {
		v := &Validator{}
		got, err := v.SetDefaultsAndValidate(tt.input)
		assert.Nil(t, err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s\nWant: %+v\nGot: %+v", tt.name, tt.want, got)
		}
	}
}
