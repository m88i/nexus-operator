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

	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"

	"github.com/m88i/nexus-operator/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var requiredObjects = []WatchedObjects{
	{
		GroupVersion: routev1.GroupVersion,
		AddToScheme:  routev1.Install,
		Objects:      []runtime.Object{&routev1.Route{}},
	},
	{
		GroupVersion: networkingv1beta1.SchemeGroupVersion,
		AddToScheme:  networkingv1beta1.AddToScheme,
		Objects:      []runtime.Object{&networkingv1beta1.Ingress{}},
	},
	{Objects: []runtime.Object{&corev1.Service{}, &appsv1.Deployment{}, &corev1.PersistentVolumeClaim{}, &corev1.ServiceAccount{}}},
}

// K8s < 3.14
func Test_controllerWatcher_WatchWithoutIngressOnKubernetes(t *testing.T) {
	cli := test.NewFakeClientBuilder().Build()
	controller := test.NewController()
	manager := test.NewManager(cli)

	watcher := NewControllerWatcher(cli, manager, controller, &v1alpha1.Nexus{})
	assert.NotNil(t, watcher)

	err := watcher.Watch(requiredObjects...)
	assert.NoError(t, err)
	// We're not watching Routes and Ingresses
	assert.False(t, watcher.AreAllObjectsWatched())
	// we should only have the objects from Kubernetes core
	assert.Len(t, controller.GetWatchedSources(), 4)
}

// K8s > 3.14
func Test_controllerWatcher_WatchWithIngressOnKubernetes(t *testing.T) {
	cli := test.NewFakeClientBuilder().WithIngress().Build()
	controller := test.NewController()
	manager := test.NewManager(cli)

	watcher := NewControllerWatcher(cli, manager, controller, &v1alpha1.Nexus{})
	assert.NotNil(t, watcher)

	err := watcher.Watch(requiredObjects...)
	assert.NoError(t, err)
	// We're not watching Routes
	assert.False(t, watcher.AreAllObjectsWatched())
	// we should only have the objects from Kubernetes core and Ingress (networking/v1beta1)
	assert.Len(t, controller.GetWatchedSources(), 5)
}

// OCP
func Test_controllerWatcher_WatchWithRouteOnOpenShift(t *testing.T) {
	cli := test.NewFakeClientBuilder().OnOpenshift().Build()
	controller := test.NewController()
	manager := test.NewManager(cli)

	watcher := NewControllerWatcher(cli, manager, controller, &v1alpha1.Nexus{})
	assert.NotNil(t, watcher)

	err := watcher.Watch(requiredObjects...)
	assert.NoError(t, err)
	// We're not watching Ingresses
	assert.False(t, watcher.AreAllObjectsWatched())
	// we should only have the objects from Kubernetes core and Route (route/v1)
	assert.Len(t, controller.GetWatchedSources(), 5)
}
