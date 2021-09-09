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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/m88i/nexus-operator/pkg/test"
)

func TestIsIngressAvailable(t *testing.T) {
	cli = test.NewFakeClientBuilder().Build()
	ingressAvailable, err := IsIngressAvailable()
	assert.Nil(t, err)
	assert.False(t, ingressAvailable)

	cli = test.NewFakeClientBuilder().WithIngress().Build()
	ingressAvailable, err = IsIngressAvailable()
	assert.Nil(t, err)
	assert.True(t, ingressAvailable)
}
