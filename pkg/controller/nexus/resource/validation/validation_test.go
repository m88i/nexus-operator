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

package validation

import (
	"fmt"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"reflect"
	"testing"
)

var (
	allDefaultsCommunityNexus = &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{Name: "nexus-test"},
		Spec: v1alpha1.NexusSpec{
			ServiceAccountName: "nexus-test",
			Resources:          DefaultResources,
			Image:              NexusCommunityLatestImage,
			LivenessProbe:      DefaultProbe.DeepCopy(),
			ReadinessProbe:     DefaultProbe.DeepCopy(),
		},
	}

	allDefaultsRedHatNexus = func() *v1alpha1.Nexus {
		nexus := allDefaultsCommunityNexus.DeepCopy()
		nexus.Spec.UseRedHatImage = true
		nexus.Spec.Image = NexusCertifiedLatestImage
		return nexus
	}()
)

func TestNewValidator(t *testing.T) {
	tests := []struct {
		name string
		disc discovery.DiscoveryInterface
		want *Validator
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
		got, err := NewValidator(tt.disc)
		assert.Nil(t, err)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s\nWant: %+v\nGot: %+v", tt.name, tt.want, got)
		}
	}

	// now let's test out error
	disc := test.NewFakeClientBuilder().Build()
	errString := "test error"
	disc.SetMockErrorForOneRequest(fmt.Errorf(errString))
	_, err := NewValidator(disc)
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
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Resources = corev1.ResourceRequirements{}
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.useRedHatImage' set to true and 'spec.image' not left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.UseRedHatImage = true
				return nexus
			}(),
			allDefaultsRedHatNexus,
		}, {
			"'spec.useRedHatImage' set to false and 'spec.image' left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Image = ""
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.successThreshold' not equal to 1",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe.SuccessThreshold = 2
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.*' and 'spec.readinessProbe.*' don't meet minimum values",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
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
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = minimumDefaultProbe.DeepCopy()
				nexus.Spec.ReadinessProbe = minimumDefaultProbe.DeepCopy()
				return nexus
			}(),
		},
		{
			"Unset 'spec.livenessProbe' and 'spec.readinessProbe'",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = nil
				nexus.Spec.ReadinessProbe = nil
				return nexus
			}(),
			allDefaultsCommunityNexus.DeepCopy(),
		},
		{
			"Invalid 'spec.imagePullPolicy'",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ImagePullPolicy = "invalid"
				return nexus
			}(),
			allDefaultsCommunityNexus.DeepCopy(),
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
				n := allDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *v1alpha1.Nexus {
				n := allDefaultsCommunityNexus.DeepCopy()
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
				n := allDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *v1alpha1.Nexus {
				n := allDefaultsCommunityNexus.DeepCopy()
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
				n := allDefaultsCommunityNexus.DeepCopy()
				n.Spec.Networking.Expose = true
				return n
			}(),
			func() *v1alpha1.Nexus {
				n := allDefaultsCommunityNexus.DeepCopy()
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
				n := allDefaultsCommunityNexus.DeepCopy()
				n.Spec.Persistence.Persistent = true
				return n
			}(),
			func() *v1alpha1.Nexus {
				n := allDefaultsCommunityNexus.DeepCopy()
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
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ServiceAccountName = ""
				return nexus
			}(),
			allDefaultsCommunityNexus,
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