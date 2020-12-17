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

package test

import (
	"reflect"
	"testing"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestContainsType(t *testing.T) {
	resources := []resource.KubernetesResource{&corev1.ServiceAccount{}}
	assert.True(t, ContainsType(resources, reflect.TypeOf(&corev1.ServiceAccount{})))
	assert.False(t, ContainsType(resources, reflect.TypeOf(&corev1.Service{})))
}

func TestEventExists(t *testing.T) {
	testReason := "reason"
	testEvent := &corev1.Event{Reason: testReason}
	c := NewFakeClient(testEvent)

	assert.False(t, EventExists(c, "some other reason"))
	assert.True(t, EventExists(c, testReason))
}

type foo interface {
	bar()
}
type concrete struct{}

func (*concrete) bar() {}
func TestIsInterfaceValueNil(t *testing.T) {
	var f foo
	assert.True(t, IsInterfaceValueNil(f))

	var c *concrete
	f = c
	assert.True(t, IsInterfaceValueNil(f))

	f = &concrete{}
	assert.False(t, IsInterfaceValueNil(f))
}
