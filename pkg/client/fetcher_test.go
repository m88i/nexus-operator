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

package client

import (
	goerrors "errors"
	"testing"

	"github.com/m88i/nexus-operator/pkg/framework"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/pkg/framework/kind"
)

func TestFetchDeployedResources(t *testing.T) {
	meta := metav1.ObjectMeta{Name: "test", Namespace: t.Name()}
	deployment := &appsv1.Deployment{ObjectMeta: meta}
	service := &corev1.Service{ObjectMeta: meta}
	managedObjectsRef := map[string]resource.KubernetesResource{
		kind.ServiceKind:    &corev1.Service{},
		kind.DeploymentKind: &appsv1.Deployment{},
		// we won't have a SA, but this is useful to test no error is triggered when a resource isn't found
		kind.SvcAccountKind: &corev1.ServiceAccount{},
	}

	cli := NewFakeClient(deployment, service)
	gotResources, err := FetchDeployedResources(managedObjectsRef, framework.Key(deployment), cli)
	assert.Nil(t, err)
	assert.Len(t, gotResources, 2)
}

func TestFetchDeployedResourcesFailure(t *testing.T) {
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "dep-test", Namespace: t.Name()}}
	// managedObjectsRef cannot be empty in order to raise error, the content is irrelevant though
	managedObjectsRef := map[string]resource.KubernetesResource{kind.DeploymentKind: &appsv1.Deployment{}}
	cli := NewFakeClient()
	mockErrorMsg := "mock error"

	cli.SetMockError(goerrors.New(mockErrorMsg))
	_, err := FetchDeployedResources(managedObjectsRef, framework.Key(dep), cli)

	assert.Contains(t, err.Error(), mockErrorMsg)
}

func TestFetch(t *testing.T) {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deployment", Namespace: t.Name()}}
	cli := NewFakeClient(deployment)
	err := Fetch(cli, framework.Key(deployment), deployment, kind.DeploymentKind)
	assert.NoError(t, err)
}

func TestNotFoundFetch(t *testing.T) {
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deployment", Namespace: t.Name()}}
	cli := NewFakeClient()
	err := Fetch(cli, framework.Key(deployment), deployment, kind.DeploymentKind)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}
