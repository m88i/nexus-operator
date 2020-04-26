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
	"testing"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
	obuildv1 "github.com/openshift/api/build/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_controllerWatcher_WatchWithOCPObjectsOnKubernetes(t *testing.T) {
	cli := test.NewFakeDiscoveryClient(false)
	controller := test.NewController()
	manager := test.NewManager()
	requiredObjects := []WatchedObjects{
		{
			GroupVersion: obuildv1.GroupVersion,
			AddToScheme:  obuildv1.Install,
			Objects:      []runtime.Object{&obuildv1.BuildConfig{}},
		},
		{
			Objects: []runtime.Object{&corev1.Service{}, &corev1.ConfigMap{}},
		},
	}

	watcher := NewControllerWatcher(cli, manager, controller, &v1alpha1.Nexus{})
	assert.NotNil(t, watcher)

	err := watcher.Watch(requiredObjects...)
	assert.NoError(t, err)
	// we are not on OpenShift
	assert.False(t, watcher.AreAllObjectsWatched())
	// we should only have the objects from Kubernetes core
	assert.Len(t, controller.GetWatchedSources(), 2)
}

func Test_controllerWatcher_WatchWithOCPObjectsOnOpenShift(t *testing.T) {
	cli := test.NewFakeDiscoveryClient(true)
	controller := test.NewController()
	manager := test.NewManager()
	requiredObjects := []WatchedObjects{
		{
			GroupVersion: obuildv1.GroupVersion,
			AddToScheme:  obuildv1.Install,
			Objects:      []runtime.Object{&obuildv1.BuildConfig{}},
		},
		{
			Objects: []runtime.Object{&corev1.Service{}, &corev1.ConfigMap{}},
		},
	}

	watcher := NewControllerWatcher(cli, manager, controller, &v1alpha1.Nexus{})
	assert.NotNil(t, watcher)

	err := watcher.Watch(requiredObjects...)
	assert.NoError(t, err)
	// we are on OpenShift
	assert.True(t, watcher.AreAllObjectsWatched())
	// we should only have the objects from Kubernetes core
	assert.Len(t, controller.GetWatchedSources(), 3)
}
