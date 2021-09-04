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
	"context"

	openapi_v2 "github.com/googleapis/gnostic/openapiv2"
	routev1 "github.com/openshift/api/route/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	discfake "k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	clienttesting "k8s.io/client-go/testing"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/framework/kind"
	"github.com/m88i/nexus-operator/pkg/util"
)

const (
	openshiftGroupVersion = "openshift.io/v1"
)

// FakeClientBuilder allows building a FakeClient according to
// the desired cluster capabilities
type FakeClientBuilder struct {
	initObjs  []runtime.Object
	scheme    *runtime.Scheme
	resources []*metav1.APIResourceList
}

// NewFakeClientBuilder will create a new fake client that is aware of minimal resource types
// and stores initObjs for initialization later
func NewFakeClientBuilder(initObjs ...runtime.Object) *FakeClientBuilder {
	s := scheme.Scheme
	util.Must(minimumSchemeBuilder().AddToScheme(s))
	res := []*metav1.APIResourceList{{GroupVersion: v1alpha1.GroupVersion.String()}}

	return &FakeClientBuilder{
		initObjs:  initObjs,
		scheme:    s,
		resources: res,
	}
}

// OnOpenshift makes the fake client aware of resources from Openshift
func (b *FakeClientBuilder) OnOpenshift() *FakeClientBuilder {
	util.Must(schemeBuilderOnOCP().AddToScheme(b.scheme))
	b.resources = append(b.resources,
		&metav1.APIResourceList{GroupVersion: openshiftGroupVersion},
		&metav1.APIResourceList{GroupVersion: routev1.GroupVersion.String(), APIResources: []metav1.APIResource{{Kind: kind.RouteKind}}})
	return b
}

// WithIngress makes the fake client aware of v1 Ingresses
func (b *FakeClientBuilder) WithIngress() *FakeClientBuilder {
	util.Must(schemeBuilderWithIngress().AddToScheme(b.scheme))
	b.resources = append(b.resources, &metav1.APIResourceList{GroupVersion: networkingv1.SchemeGroupVersion.String(), APIResources: []metav1.APIResource{{Kind: kind.IngressKind}}})
	return b
}

// Build returns the fake discovery client
func (b *FakeClientBuilder) Build() *FakeClient {
	return &FakeClient{
		scheme: b.scheme,
		client: fake.NewFakeClientWithScheme(b.scheme, b.initObjs...),
		disc: &discfake.FakeDiscovery{
			Fake: &clienttesting.Fake{
				Resources: b.resources,
			},
		},
	}
}

func minimumSchemeBuilder() *runtime.SchemeBuilder {
	return &runtime.SchemeBuilder{v1alpha1.SchemeBuilder.AddToScheme}
}

func schemeBuilderOnOCP() *runtime.SchemeBuilder {
	return &runtime.SchemeBuilder{routev1.Install}
}

func schemeBuilderWithIngress() *runtime.SchemeBuilder {
	return &runtime.SchemeBuilder{networkingv1.AddToScheme}
}

// FakeClient wraps an API fake client to allow mocked error responses
// Useful for covering errors other than NotFound
// It also wraps a fake discovery client
// FakeClient implements both client.Client and discovery.DiscoveryInterface
type FakeClient struct {
	scheme           *runtime.Scheme
	client           client.Client
	disc             discovery.DiscoveryInterface
	mockErr          error
	shouldClearError bool
}

// Scheme returns the fake client's scheme
func (c *FakeClient) Scheme() *runtime.Scheme {
	return c.scheme
}

// SetMockError sets the error which should be returned for the following requests
// This error will continue to be served until cleared with c.ClearMockError()
func (c *FakeClient) SetMockError(err error) {
	c.shouldClearError = false
	c.mockErr = err
}

// SetMockErrorForOneRequest sets the error which should be returned for the following request
// this error will be set to nil after the next request
func (c *FakeClient) SetMockErrorForOneRequest(err error) {
	c.shouldClearError = true
	c.mockErr = err
}

// ClearMockError unsets any mock errors previously set
func (c *FakeClient) ClearMockError() {
	c.shouldClearError = false
	c.mockErr = nil
}

func (c FakeClient) RESTClient() rest.Interface {
	return c.disc.RESTClient()
}

func (c *FakeClient) ServerGroups() (*metav1.APIGroupList, error) {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return nil, c.mockErr
	}
	return c.disc.ServerGroups()
}

func (c *FakeClient) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return nil, c.mockErr
	}
	return c.disc.ServerResourcesForGroupVersion(groupVersion)
}

// Deprecated: use ServerGroupsAndResources instead.
func (c *FakeClient) ServerResources() ([]*metav1.APIResourceList, error) {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return nil, c.mockErr
	}
	return c.disc.ServerResources()
}

func (c *FakeClient) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return nil, nil, c.mockErr
	}
	return c.disc.ServerGroupsAndResources()
}

func (c *FakeClient) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return nil, c.mockErr
	}
	return c.disc.ServerPreferredResources()
}

func (c *FakeClient) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return nil, c.mockErr
	}
	return c.disc.ServerPreferredNamespacedResources()
}

func (c *FakeClient) ServerVersion() (*version.Info, error) {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return nil, c.mockErr
	}
	return c.disc.ServerVersion()
}

func (c *FakeClient) OpenAPISchema() (*openapi_v2.Document, error) {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return nil, c.mockErr
	}
	return c.disc.OpenAPISchema()
}

func (c *FakeClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return c.mockErr
	}
	return c.client.Get(ctx, key, obj)
}

func (c *FakeClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return c.mockErr
	}
	return c.client.List(ctx, list, opts...)
}

func (c *FakeClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return c.mockErr
	}
	return c.client.Create(ctx, obj, opts...)
}

func (c *FakeClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return c.mockErr
	}
	return c.client.Delete(ctx, obj, opts...)
}

func (c *FakeClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return c.mockErr
	}
	return c.client.Update(ctx, obj, opts...)
}

func (c *FakeClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return c.mockErr
	}
	return c.client.Patch(ctx, obj, patch, opts...)
}

func (c *FakeClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	if c.mockErr != nil {
		if c.shouldClearError {
			defer c.ClearMockError()
		}
		return c.mockErr
	}
	return c.client.DeleteAllOf(ctx, obj, opts...)
}

func (c FakeClient) Status() client.StatusWriter {
	return c.client.Status()
}
