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
	"reflect"
	"strings"
	"testing"

	"github.com/m88i/nexus-operator/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/cluster/openshift"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const testErrorMsg = "test"

func TestNewFakeClientBuilder(t *testing.T) {
	nexus := &v1alpha1.Nexus{}
	b := NewFakeClientBuilder(nexus)

	// client.Client
	assert.Len(t, b.scheme.KnownTypes(v1alpha1.GroupVersion), 10)
	assert.Contains(t, b.scheme.KnownTypes(v1alpha1.GroupVersion), strings.Split(reflect.TypeOf(&v1alpha1.Nexus{}).String(), ".")[1])
	assert.Contains(t, b.scheme.KnownTypes(v1alpha1.GroupVersion), strings.Split(reflect.TypeOf(&v1alpha1.NexusList{}).String(), ".")[1])

	// discovery.DiscoveryInterface
	assert.True(t, resourceListsContainsGroupVersion(b.resources, v1alpha1.GroupVersion.String()))

	// initObjs
	assert.Len(t, b.initObjs, 1)
	assert.Contains(t, b.initObjs, nexus)
}

func TestFakeClientBuilder_OnOpenshift(t *testing.T) {
	b := NewFakeClientBuilder().OnOpenshift()

	// client.Client
	assert.Len(t, b.scheme.KnownTypes(routev1.GroupVersion), 10)
	assert.Contains(t, b.scheme.KnownTypes(routev1.GroupVersion), strings.Split(reflect.TypeOf(&routev1.Route{}).String(), ".")[1])
	assert.Contains(t, b.scheme.KnownTypes(routev1.GroupVersion), strings.Split(reflect.TypeOf(&routev1.RouteList{}).String(), ".")[1])

	// discovery.DiscoveryInterface
	assert.True(t, resourceListsContainsGroupVersion(b.resources, routev1.GroupVersion.String()))
	assert.True(t, resourceListsContainsGroupVersion(b.resources, openshiftGroupVersion))
}

func TestFakeClientBuilder_WithIngress(t *testing.T) {
	b := NewFakeClientBuilder().WithIngress()

	// client.Client
	assert.Len(t, b.scheme.KnownTypes(networkingv1beta1.SchemeGroupVersion), 12)
	assert.Contains(t, b.scheme.KnownTypes(networkingv1beta1.SchemeGroupVersion), strings.Split(reflect.TypeOf(&networkingv1beta1.Ingress{}).String(), ".")[1])
	assert.Contains(t, b.scheme.KnownTypes(networkingv1beta1.SchemeGroupVersion), strings.Split(reflect.TypeOf(&networkingv1beta1.IngressList{}).String(), ".")[1])

	// discovery.DiscoveryInterface
	assert.True(t, resourceListsContainsGroupVersion(b.resources, networkingv1beta1.SchemeGroupVersion.String()))
}

func TestFakeClientBuilder_Build(t *testing.T) {
	nexus := &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "nexus"}}
	route := &routev1.Route{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "route"}}
	ingress := &networkingv1beta1.Ingress{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "ingress"}}

	b := NewFakeClientBuilder(nexus)
	c := b.Build()
	assert.NotNil(t, c.disc)
	assert.NotNil(t, c.client)
	assert.NoError(t, c.client.Get(ctx.TODO(), client.ObjectKey{
		Namespace: nexus.Namespace,
		Name:      nexus.Name,
	}, nexus))
	ocp, _ := openshift.IsOpenShift(c)
	assert.False(t, ocp)
	withRoute, _ := openshift.IsRouteAvailable(c)
	assert.False(t, withRoute)
	withIngress, _ := kubernetes.IsIngressAvailable(c)
	assert.False(t, withIngress)

	// on Openshift
	b = NewFakeClientBuilder(nexus, route).OnOpenshift()
	c = b.Build()
	assert.NoError(t, c.client.Get(ctx.TODO(), client.ObjectKey{
		Namespace: route.Namespace,
		Name:      route.Name,
	}, route))
	ocp, _ = openshift.IsOpenShift(c)
	assert.True(t, ocp)
	withRoute, _ = openshift.IsRouteAvailable(c)
	assert.True(t, withRoute)
	withIngress, _ = kubernetes.IsIngressAvailable(c)
	assert.False(t, withIngress)

	// with Ingress
	b = NewFakeClientBuilder(nexus, ingress).WithIngress()
	c = b.Build()
	assert.NoError(t, c.client.Get(ctx.TODO(), client.ObjectKey{
		Namespace: ingress.Namespace,
		Name:      ingress.Name,
	}, ingress))
	ocp, _ = openshift.IsOpenShift(c)
	assert.False(t, ocp)
	withRoute, _ = openshift.IsRouteAvailable(c)
	assert.False(t, withRoute)
	withIngress, _ = kubernetes.IsIngressAvailable(c)
	assert.True(t, withIngress)
}

func resourceListsContainsGroupVersion(lists []*metav1.APIResourceList, gv string) bool {
	for _, list := range lists {
		if list.GroupVersion == gv {
			return true
		}
	}
	return false
}

func TestFakeClient_SetMockError(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	assert.Nil(t, c.mockErr)
	assert.False(t, c.shouldClearError)
	c.SetMockError(fmt.Errorf(testErrorMsg))
	assert.Equal(t, c.mockErr.Error(), testErrorMsg)
	assert.False(t, c.shouldClearError)
}

func TestFakeClient_SetMockErrorForOneRequest(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	assert.Nil(t, c.mockErr)
	assert.False(t, c.shouldClearError)
	c.SetMockErrorForOneRequest(fmt.Errorf(testErrorMsg))
	assert.Equal(t, c.mockErr.Error(), testErrorMsg)
	assert.True(t, c.shouldClearError)
}

func TestFakeClient_ClearMockError(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	assert.Nil(t, c.mockErr)
	assert.False(t, c.shouldClearError)
	c.SetMockErrorForOneRequest(fmt.Errorf(testErrorMsg))
	assert.Equal(t, c.mockErr.Error(), testErrorMsg)
	assert.True(t, c.shouldClearError)
	c.ClearMockError()
	assert.Nil(t, c.mockErr)
	assert.False(t, c.shouldClearError)
}

func TestFakeClient_RESTClient(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	assert.Equal(t, c.disc.RESTClient(), c.RESTClient())
}

func TestFakeClient_ServerGroups(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	_, err := c.ServerGroups()
	assert.Equal(t, mockErr, err)

	want, wantErr := c.disc.ServerGroups()
	got, gotErr := c.ServerGroups()
	assert.Equal(t, want, got)
	assert.Equal(t, wantErr, gotErr)
	assert.NotEqual(t, gotErr, mockErr)
}

func TestFakeClient_ServerResourcesForGroupVersion(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	_, err := c.ServerResourcesForGroupVersion("")
	assert.Equal(t, mockErr, err)

	want, wantErr := c.disc.ServerResourcesForGroupVersion("")
	got, gotErr := c.ServerResourcesForGroupVersion("")
	assert.Equal(t, want, got)
	assert.Equal(t, wantErr, gotErr)
	assert.NotEqual(t, gotErr, mockErr)
}

// Deprecated: use ServerGroupsAndResources instead.
func TestFakeClient_ServerResources(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	_, err := c.ServerResources()
	assert.Equal(t, mockErr, err)

	want, wantErr := c.disc.ServerResources()
	got, gotErr := c.ServerResources()
	assert.Equal(t, want, got)
	assert.Equal(t, wantErr, gotErr)
	assert.NotEqual(t, gotErr, mockErr)
}

func TestFakeClient_ServerGroupsAndResources(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	_, _, err := c.ServerGroupsAndResources()
	assert.Equal(t, mockErr, err)

	want1, want2, wantErr := c.disc.ServerGroupsAndResources()
	got1, got2, gotErr := c.ServerGroupsAndResources()
	assert.Equal(t, want1, got1)
	assert.Equal(t, want2, got2)
	assert.Equal(t, wantErr, gotErr)
	assert.NotEqual(t, gotErr, mockErr)
}

func TestFakeClient_ServerPreferredResources(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	_, err := c.ServerPreferredResources()
	assert.Equal(t, mockErr, err)

	want, wantErr := c.disc.ServerPreferredResources()
	got, gotErr := c.ServerPreferredResources()
	assert.Equal(t, want, got)
	assert.Equal(t, wantErr, gotErr)
	assert.NotEqual(t, gotErr, mockErr)
}

func TestFakeClient_ServerPreferredNamespacedResources(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	_, err := c.ServerPreferredNamespacedResources()
	assert.Equal(t, mockErr, err)

	want, wantErr := c.disc.ServerPreferredNamespacedResources()
	got, gotErr := c.ServerPreferredNamespacedResources()
	assert.Equal(t, want, got)
	assert.Equal(t, wantErr, gotErr)
	assert.NotEqual(t, gotErr, mockErr)
}

func TestFakeClient_ServerVersion(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	_, err := c.ServerVersion()
	assert.Equal(t, mockErr, err)

	want, wantErr := c.disc.ServerVersion()
	got, gotErr := c.ServerVersion()
	assert.Equal(t, want, got)
	assert.Equal(t, wantErr, gotErr)
	assert.NotEqual(t, gotErr, mockErr)
}

func TestFakeClient_OpenAPISchema(t *testing.T) {
	c := NewFakeClientBuilder().Build()
	mockErr := fmt.Errorf(testErrorMsg)
	c.SetMockErrorForOneRequest(mockErr)
	_, err := c.OpenAPISchema()
	assert.Equal(t, mockErr, err)

	want, wantErr := c.disc.OpenAPISchema()
	got, gotErr := c.OpenAPISchema()
	assert.Equal(t, want, got)
	assert.Equal(t, wantErr, gotErr)
	assert.NotEqual(t, gotErr, mockErr)
}

func TestFakeClient_Get(t *testing.T) {
	c := NewFakeClientBuilder().Build()
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
	c := NewFakeClientBuilder().Build()
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
	c := NewFakeClientBuilder().Build()
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
	c := NewFakeClientBuilder().Build()
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
	c := NewFakeClientBuilder().Build()
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
	c := NewFakeClientBuilder().Build()
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
	c := NewFakeClientBuilder().Build()
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
	c := NewFakeClientBuilder().Build()
	assert.Equal(t, c.client.Status(), c.Status())
}
