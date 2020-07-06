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
	"reflect"
	"testing"
)

func TestManager_validate(t *testing.T) {
	tests := []struct {
		name             string
		ocp              bool
		routeAvailable   bool
		ingressAvailable bool
		input            v1alpha1.Nexus
		wantError        bool
	}{
		{
			"Valid Nexus with Ingress on K8s",
			false,
			false,
			true,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com"}}},
			false,
		},
		{
			"Valid Nexus with Ingress and TLS secret on K8s",
			false,
			false,
			true,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com", TLS: v1alpha1.NexusNetworkingTLS{SecretName: "tt-secret"}}}},
			false,
		},
		{
			"Valid Nexus with Ingress on K8s, but Ingress unavailable (Kubernetes < 1.14)",
			false,
			false,
			false,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com"}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on K8s and no 'spec.networking.host'",
			false,
			false,
			true,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType}}},
			true,
		},
		{
			"Invalid Nexus with Ingress on K8s and 'spec.networking.mandatory' set to 'true'",
			false,
			false,
			true,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com", TLS: v1alpha1.NexusNetworkingTLS{Mandatory: true}}}},
			true,
		},
		{
			"Invalid Nexus with Route on K8s",
			false,
			false,
			true,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.RouteExposeType}}},
			true,
		},
		{
			"Valid Nexus with Route on OCP",
			true,
			true,
			false,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.RouteExposeType}}},
			false,
		},
		{
			"Valid Nexus with Route on OCP with mandatory TLS",
			true,
			true,
			false,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.RouteExposeType, TLS: v1alpha1.NexusNetworkingTLS{Mandatory: true}}}},
			false,
		},
		{
			"Invalid Nexus with Ingress on OCP",
			true,
			true,
			false,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType, Host: "example.com"}}},
			true,
		},
		{
			"Valid Nexus with Node Port",
			false, // unimportant
			false, // unimportant
			false, // unimportant
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.NodePortExposeType, NodePort: 8080}}},
			false,
		},
		{
			"Invalid Nexus with Node Port and no port informed",
			false, // unimportant
			false, // unimportant
			false, // unimportant
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.NodePortExposeType}}},
			true,
		},
	}

	for _, tt := range tests {
		manager := &Manager{
			nexus:            &tt.input,
			routeAvailable:   tt.routeAvailable,
			ingressAvailable: tt.ingressAvailable,
			ocp:              tt.ocp,
		}
		if err := manager.validate(); (err != nil) != tt.wantError {
			t.Errorf("TestManager_validate() - %s\nWantError: %v\tError: %v", tt.name, tt.wantError, err)
		}
	}
}

func TestNewManager_setDefaults(t *testing.T) {
	tests := []struct {
		name             string
		ocp              bool
		routeAvailable   bool
		ingressAvailable bool
		input            v1alpha1.Nexus
		want             v1alpha1.Nexus
	}{
		{
			"'spec.networking.exposeAs' left blank on OCP",
			true,
			true,
			false, // unimportant
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true}}},
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.RouteExposeType}}},
		},
		{
			"'spec.networking.exposeAs' left blank on K8s",
			false,
			false,
			true,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true}}},
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.IngressExposeType}}},
		},
		{
			"'spec.networking.exposeAs' left blank on K8s, but Ingress unavailable",
			false,
			false,
			false,
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true}}},
			v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Networking: v1alpha1.NexusNetworking{Expose: true, ExposeAs: v1alpha1.NodePortExposeType}}},
		},
	}

	for _, tt := range tests {
		manager := &Manager{
			nexus:            &tt.input,
			routeAvailable:   tt.routeAvailable,
			ingressAvailable: tt.ingressAvailable,
			ocp:              tt.ocp,
		}
		manager.setDefaults()
		if !reflect.DeepEqual(*manager.nexus, tt.want) {
			t.Errorf("TestManager_setDefaults() - %s\nWant: %v\tGot: %v", tt.name, tt.want, *manager.nexus)
		}
	}
}
