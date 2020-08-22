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

package framework

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	deployment := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: "deployment", Namespace: t.Name()}}
	cli := test.NewFakeClientBuilder(deployment).Build()
	err := Fetch(cli, Key(deployment), deployment)
	assert.NoError(t, err)
}

func TestNotFoundFetch(t *testing.T) {
	deployment := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: "deployment", Namespace: t.Name()}}
	cli := test.NewFakeClientBuilder().Build()
	err := Fetch(cli, Key(deployment), deployment)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}
