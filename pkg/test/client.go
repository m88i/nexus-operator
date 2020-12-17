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

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/m88i/nexus-operator/api/v1alpha1"
)

// FakeClient wraps an API fake client to allow mocked error responses
type FakeClient struct {
	client           client.Client
	mockErr          error
	shouldClearError bool
}

// NewFakeClient returns a FakeClient.
// You may initialize it with a slice of runtime.Object.
func NewFakeClient(initObjs ...runtime.Object) *FakeClient {
	return &FakeClient{client: fake.NewFakeClientWithScheme(scheme(), initObjs...)}
}

func scheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(v1alpha1.AddToScheme(s))
	utilruntime.Must(routev1.Install(s))
	return s
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
