// Copyright 2020 Red Hat, Inc. and/or its affiliates
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
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// NewController ...
func NewController() MockController {
	return &controllerMock{}
}

// MockController ...
type MockController interface {
	controller.Controller
	// GetWatchedSources gets all the sources added to the Mock
	GetWatchedSources() []source.Source
}

// controllerMock mock to handle with operator runtime internals
type controllerMock struct {
	watches []source.Source
}

func (c *controllerMock) Reconcile(reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (c *controllerMock) Watch(src source.Source, eventhandler handler.EventHandler, predicates ...predicate.Predicate) error {
	c.watches = append(c.watches, src)
	return nil
}

func (c *controllerMock) Start(stop <-chan struct{}) error {
	return nil
}

func (c *controllerMock) GetWatchedSources() []source.Source {
	return c.watches
}

// NewManager ...
func NewManager() manager.Manager {
	return &managerMock{scheme: runtime.NewScheme()}
}

type managerMock struct {
	scheme *runtime.Scheme
}

func (m managerMock) AddMetricsExtraHandler(path string, handler http.Handler) error {
	panic("implement me")
}

func (m managerMock) Elected() <-chan struct{} {
	panic("implement me")
}

func (m managerMock) Add(manager.Runnable) error {
	panic("implement me")
}

func (m managerMock) SetFields(interface{}) error {
	panic("implement me")
}

func (m managerMock) AddHealthzCheck(name string, check healthz.Checker) error {
	panic("implement me")
}

func (m managerMock) AddReadyzCheck(name string, check healthz.Checker) error {
	panic("implement me")
}

func (m managerMock) Start(<-chan struct{}) error {
	panic("implement me")
}

func (m managerMock) GetConfig() *rest.Config {
	panic("implement me")
}

func (m managerMock) GetScheme() *runtime.Scheme {
	return m.scheme
}

func (m managerMock) GetClient() client.Client {
	panic("implement me")
}

func (m managerMock) GetFieldIndexer() client.FieldIndexer {
	panic("implement me")
}

func (m managerMock) GetCache() cache.Cache {
	panic("implement me")
}

func (m managerMock) GetEventRecorderFor(name string) record.EventRecorder {
	panic("implement me")
}

func (m managerMock) GetRESTMapper() meta.RESTMapper {
	panic("implement me")
}

func (m managerMock) GetAPIReader() client.Reader {
	panic("implement me")
}

func (m managerMock) GetWebhookServer() *webhook.Server {
	panic("implement me")
}
