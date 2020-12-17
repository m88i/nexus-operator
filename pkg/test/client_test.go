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
	ctx "context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/framework"
)

const testErrorMsg = "test"

func TestNewFakeClient(t *testing.T) {
	dep := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "dep", Namespace: t.Name()}}
	key := framework.Key(dep)
	c := NewFakeClient(dep)
	assert.False(t, IsInterfaceValueNil(c.client)) // make sure there actually is a client
	assert.NoError(t, c.Get(ctx.Background(), key, dep))
}

func TestFakeClient_Get(t *testing.T) {
	c := NewFakeClient()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	err := c.Get(ctx.TODO(), client.ObjectKey{}, &v1alpha1.Nexus{})
	assert.Equal(t, mockErr, err)

	want := c.client.Get(ctx.TODO(), client.ObjectKey{}, &v1alpha1.Nexus{})
	got := c.Get(ctx.TODO(), client.ObjectKey{}, &v1alpha1.Nexus{})
	assert.Equal(t, want, got)
	assert.NotEqual(t, got, mockErr)
}

func TestFakeClient_List(t *testing.T) {
	c := NewFakeClient()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	err := c.List(ctx.TODO(), &v1alpha1.NexusList{})
	assert.Equal(t, mockErr, err)

	want := c.client.List(ctx.TODO(), &v1alpha1.NexusList{})
	got := c.List(ctx.TODO(), &v1alpha1.NexusList{})
	assert.Equal(t, want, got)
	assert.NotEqual(t, got, mockErr)
}

func TestFakeClient_Create(t *testing.T) {
	c := NewFakeClient()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	err := c.Create(ctx.TODO(), &v1alpha1.Nexus{})
	assert.Equal(t, mockErr, err)

	want := c.client.Create(ctx.TODO(), &v1alpha1.Nexus{})
	got := c.Create(ctx.TODO(), &v1alpha1.Nexus{})
	assert.Equal(t, want, got)
	assert.NotEqual(t, got, mockErr)
}

func TestFakeClient_Delete(t *testing.T) {
	c := NewFakeClient()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	err := c.Delete(ctx.TODO(), &v1alpha1.Nexus{})
	assert.Equal(t, mockErr, err)

	want := c.client.Delete(ctx.TODO(), &v1alpha1.Nexus{})
	got := c.Delete(ctx.TODO(), &v1alpha1.Nexus{})
	assert.Equal(t, want, got)
	assert.NotEqual(t, got, mockErr)
}

func TestFakeClient_Update(t *testing.T) {
	c := NewFakeClient()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	err := c.Update(ctx.TODO(), &v1alpha1.Nexus{})
	assert.Equal(t, mockErr, err)

	want := c.client.Update(ctx.TODO(), &v1alpha1.Nexus{})
	got := c.Update(ctx.TODO(), &v1alpha1.Nexus{})
	assert.Equal(t, want, got)
	assert.NotEqual(t, got, mockErr)
}

func TestFakeClient_Patch(t *testing.T) {
	c := NewFakeClient()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	err := c.Patch(ctx.TODO(), &v1alpha1.Nexus{}, client.MergeFrom(&v1alpha1.Nexus{}))
	assert.Equal(t, mockErr, err)

	want := c.Patch(ctx.TODO(), &v1alpha1.Nexus{}, client.MergeFrom(&v1alpha1.Nexus{}))
	got := c.Patch(ctx.TODO(), &v1alpha1.Nexus{}, client.MergeFrom(&v1alpha1.Nexus{}))
	assert.Equal(t, want, got)
	assert.NotEqual(t, got, mockErr)
}

func TestFakeClient_DeleteAllOf(t *testing.T) {
	c := NewFakeClient()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	err := c.DeleteAllOf(ctx.TODO(), &v1alpha1.Nexus{})
	assert.Equal(t, mockErr, err)

	want := c.client.DeleteAllOf(ctx.TODO(), &v1alpha1.Nexus{})
	got := c.DeleteAllOf(ctx.TODO(), &v1alpha1.Nexus{})
	assert.Equal(t, want, got)
	assert.NotEqual(t, got, mockErr)
}

func TestFakeClient_Status(t *testing.T) {
	c := NewFakeClient()
	assert.Equal(t, c.client.Status(), c.Status())
}
