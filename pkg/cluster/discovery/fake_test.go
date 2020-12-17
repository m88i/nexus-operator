package discovery

import (
	"fmt"
	"testing"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/pkg/framework/kind"
)

const testErrorMsg = "test"

func TestNewFakeDiscBuilder(t *testing.T) {
	b := NewFakeDiscBuilder()
	assert.True(t, resourceListsContainsGroupVersionKind(b.resources, m88iGroupVersion.String(), kind.NexusKind))
}

func TestFakeClientBuilder_OnOpenshift(t *testing.T) {
	b := NewFakeDiscBuilder().OnOpenshift()
	assert.True(t, resourceListsContainsGroupVersionKind(b.resources, routev1.GroupVersion.String(), kind.RouteKind))
	assert.True(t, resourceListsContainsGroupVersion(b.resources, openshiftGroupVersion))
}

func TestFakeClientBuilder_WithIngress(t *testing.T) {
	b := NewFakeDiscBuilder().WithIngress()
	assert.True(t, resourceListsContainsGroupVersionKind(b.resources, networkingv1.SchemeGroupVersion.String(), kind.IngressKind))
}

func TestFakeClientBuilder_WithLegacyIngress(t *testing.T) {
	b := NewFakeDiscBuilder().WithLegacyIngress()
	assert.True(t, resourceListsContainsGroupVersionKind(b.resources, networkingv1beta1.SchemeGroupVersion.String(), kind.IngressKind))
}

func TestFakeClientBuilder_Build(t *testing.T) {
	// base
	SetClient(NewFakeDiscBuilder().Build())
	ocp, _ := IsOpenShift()
	assert.False(t, ocp)
	withRoute, _ := IsRouteAvailable()
	assert.False(t, withRoute)
	withIngress, _ := IsIngressAvailable()
	assert.False(t, withIngress)
	withLegacyIngress, _ := IsLegacyIngressAvailable()
	assert.False(t, withLegacyIngress)

	// on Openshift
	SetClient(NewFakeDiscBuilder().OnOpenshift().Build())
	ocp, _ = IsOpenShift()
	assert.True(t, ocp)
	withRoute, _ = IsRouteAvailable()
	assert.True(t, withRoute)
	withIngress, _ = IsIngressAvailable()
	assert.False(t, withIngress)
	withLegacyIngress, _ = IsLegacyIngressAvailable()
	assert.False(t, withLegacyIngress)

	// with Ingress
	SetClient(NewFakeDiscBuilder().WithIngress().Build())
	ocp, _ = IsOpenShift()
	assert.False(t, ocp)
	withRoute, _ = IsRouteAvailable()
	assert.False(t, withRoute)
	withIngress, _ = IsIngressAvailable()
	assert.True(t, withIngress)
	withLegacyIngress, _ = IsLegacyIngressAvailable()
	assert.False(t, withLegacyIngress)

	// with v1beta1 Ingress
	SetClient(NewFakeDiscBuilder().WithLegacyIngress().Build())
	ocp, _ = IsOpenShift()
	assert.False(t, ocp)
	withRoute, _ = IsRouteAvailable()
	assert.False(t, withRoute)
	withIngress, _ = IsIngressAvailable()
	assert.False(t, withIngress)
	withLegacyIngress, _ = IsLegacyIngressAvailable()
	assert.True(t, withLegacyIngress)
}

func resourceListsContainsGroupVersion(lists []*metav1.APIResourceList, gv string) bool {
	for _, list := range lists {
		if list.GroupVersion == gv {
			return true
		}
	}
	return false
}

func resourceListsContainsGroupVersionKind(lists []*metav1.APIResourceList, gv, kind string) bool {
	for _, list := range lists {
		if list.GroupVersion == gv {
			for _, res := range list.APIResources {
				if res.Kind == kind {
					return true
				}
			}
			// correct group, incorrect kind
			return false
		}
	}
	return false
}

func TestFakeClient_SetMockError(t *testing.T) {
	c := NewFakeDiscBuilder().Build()
	assert.Nil(t, c.mockErr)
	assert.False(t, c.shouldClearError)
	c.SetMockError(fmt.Errorf(testErrorMsg))
	assert.Equal(t, c.mockErr.Error(), testErrorMsg)
	assert.False(t, c.shouldClearError)
}

func TestFakeClient_SetMockErrorForOneRequest(t *testing.T) {
	c := NewFakeDiscBuilder().Build()
	assert.Nil(t, c.mockErr)
	assert.False(t, c.shouldClearError)
	c.SetMockErrorForOneRequest(fmt.Errorf(testErrorMsg))
	assert.Equal(t, c.mockErr.Error(), testErrorMsg)
	assert.True(t, c.shouldClearError)
}

func TestFakeClient_ClearMockError(t *testing.T) {
	c := NewFakeDiscBuilder().Build()
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
	c := NewFakeDiscBuilder().Build()
	assert.Equal(t, c.disc.RESTClient(), c.RESTClient())
}

func TestFakeClient_ServerGroups(t *testing.T) {
	c := NewFakeDiscBuilder().Build()
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
	c := NewFakeDiscBuilder().Build()
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
	c := NewFakeDiscBuilder().Build()
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
	c := NewFakeDiscBuilder().Build()
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
	c := NewFakeDiscBuilder().Build()
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
	c := NewFakeDiscBuilder().Build()
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
	c := NewFakeDiscBuilder().Build()
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
	c := NewFakeDiscBuilder().Build()
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
