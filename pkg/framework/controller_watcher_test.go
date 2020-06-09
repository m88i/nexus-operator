//     Copyright 2020 Nexus Operator and/or its authors
//
//     This file is part of Nexus Operator.
//
//     Nexus Operator is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     Nexus Operator is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with Nexus Operator.  If not, see <https://www.gnu.org/licenses/>.

package framework

import (
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	"testing"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
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
	cli := test.NewFakeDiscoveryClient().Build()
	controller := test.NewController()
	manager := test.NewManager()

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
	cli := test.NewFakeDiscoveryClient().WithIngress().Build()
	controller := test.NewController()
	manager := test.NewManager()

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
	cli := test.NewFakeDiscoveryClient().OnOpenshift().Build()
	controller := test.NewController()
	manager := test.NewManager()

	watcher := NewControllerWatcher(cli, manager, controller, &v1alpha1.Nexus{})
	assert.NotNil(t, watcher)

	err := watcher.Watch(requiredObjects...)
	assert.NoError(t, err)
	// We're not watching Ingresses
	assert.False(t, watcher.AreAllObjectsWatched())
	// we should only have the objects from Kubernetes core and Route (route/v1)
	assert.Len(t, controller.GetWatchedSources(), 5)
}
