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

package discovery

import (
	"errors"
	"testing"

	"k8s.io/client-go/discovery"

	"github.com/stretchr/testify/assert"
)

func TestIsIngressAvailable(t *testing.T) {
	SetClient(NewFakeDiscBuilder().Build())
	ingressAvailable, err := IsIngressAvailable()
	assert.Nil(t, err)
	assert.False(t, ingressAvailable)

	SetClient(NewFakeDiscBuilder().WithIngress().Build())
	ingressAvailable, err = IsIngressAvailable()
	assert.Nil(t, err)
	assert.True(t, ingressAvailable)
}

func TestIsLegacyIngressAvailable(t *testing.T) {
	SetClient(NewFakeDiscBuilder().Build())
	ingressAvailable, err := IsLegacyIngressAvailable()
	assert.Nil(t, err)
	assert.False(t, ingressAvailable)

	SetClient(NewFakeDiscBuilder().WithLegacyIngress().Build())
	ingressAvailable, err = IsLegacyIngressAvailable()
	assert.Nil(t, err)
	assert.True(t, ingressAvailable)
}

func TestIsAnyIngressAvailable(t *testing.T) {
	testCases := []struct {
		name                 string
		disc                 discovery.DiscoveryInterface
		wantIngressAvailable bool
		wantError            bool
	}{
		{
			"no ingresses available",
			NewFakeDiscBuilder().Build(),
			false,
			false,
		},
		{
			"v1beta1 ingress only available",
			NewFakeDiscBuilder().WithLegacyIngress().Build(),
			true,
			false,
		},
		{
			"v1 ingress available",
			NewFakeDiscBuilder().WithIngress().Build(),
			true,
			false,
		},
		{
			"both ingresses available",
			NewFakeDiscBuilder().WithIngress().WithLegacyIngress().Build(),
			true,
			false,
		},
		{
			"both ingresses available. One call fails, the other succeeds",
			func() discovery.DiscoveryInterface {
				d := NewFakeDiscBuilder().WithIngress().WithLegacyIngress().Build()
				d.SetMockErrorForOneRequest(errors.New("mock err"))
				return d
			}(),
			true,
			false,
		},
		{
			"both ingresses available. Both calls fail",
			func() discovery.DiscoveryInterface {
				d := NewFakeDiscBuilder().WithIngress().WithLegacyIngress().Build()
				d.SetMockError(errors.New("mock err"))
				return d
			}(),
			false,
			true,
		},
	}

	for _, tc := range testCases {
		SetClient(tc.disc)
		ingressAvailable, err := IsAnyIngressAvailable()
		if tc.wantIngressAvailable != ingressAvailable || tc.wantError != (err != nil) {
			t.Errorf("%s\nwantIngressAvailable: %v\t got: %v\nwantError: %v\tgotError: %#v\n", tc.name, tc.wantIngressAvailable, ingressAvailable, tc.wantError, err)
		}
	}
}
