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

package server

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/aicura/nexus"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/meta"
	operatorclient "github.com/m88i/nexus-operator/pkg/client"
)

// createNewServerAndKubeCli creates a new fake server and kubernetes fake client to be used in tests for this package;
// A nexus CR instance is also added to the Fake client context.
func createNewServerAndKubeCli(t *testing.T, objects ...runtime.Object) (*server, client.Client) {
	nexusInstance := &v1alpha1.Nexus{ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()}}
	objects = append(objects, nexusInstance)
	cli := operatorclient.NewFakeClient(objects...)
	server := &server{
		nexus:     nexusInstance,
		k8sclient: cli,
		nexuscli:  nexus.NewFakeClient(),
		status:    &v1alpha1.OperationsStatus{},
	}

	return server, cli
}

func nexusAPIFakeBuilder(url, user, pass string) *nexus.Client {
	return nexus.NewFakeClient()
}

func Test_server_getNexusEndpoint(t *testing.T) {
	instance := &v1alpha1.Nexus{
		Spec:       v1alpha1.NexusSpec{},
		ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()},
	}
	svc := &corev1.Service{
		ObjectMeta: meta.DefaultObjectMeta(instance),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     80,
					TargetPort: intstr.IntOrString{
						IntVal: 8081,
					},
				},
			},
			Selector:        meta.GenerateLabels(instance),
			SessionAffinity: corev1.ServiceAffinityNone,
		},
	}
	cli := operatorclient.NewFakeClient(svc, instance)
	s := server{
		nexus:     instance,
		k8sclient: cli,
	}
	URL, err := s.getNexusEndpoint()
	assert.NoError(t, err)
	assert.NotEmpty(t, URL)
	assert.Contains(t, URL, instance.Name)
	_, err = url.Parse(URL)
	assert.NoError(t, err)
}

func Test_server_getNexusEndpointNoURL(t *testing.T) {
	instance := &v1alpha1.Nexus{
		Spec:       v1alpha1.NexusSpec{},
		ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()},
	}
	cli := operatorclient.NewFakeClient(instance)
	s := server{
		nexus:     instance,
		k8sclient: cli,
	}
	URL, err := s.getNexusEndpoint()
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
	assert.Empty(t, URL)
}

func Test_server_isServerReady(t *testing.T) {
	instance := &v1alpha1.Nexus{
		Spec:       v1alpha1.NexusSpec{},
		ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()},
		Status: v1alpha1.NexusStatus{
			DeploymentStatus: appv1.DeploymentStatus{
				AvailableReplicas: 1,
			},
		},
	}
	s := server{nexus: instance, status: &v1alpha1.OperationsStatus{}}
	assert.True(t, s.isServerReady())
}

func Test_server_serverNotReady(t *testing.T) {
	instance := &v1alpha1.Nexus{
		Spec:       v1alpha1.NexusSpec{},
		ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()},
	}
	s := server{nexus: instance, status: &v1alpha1.OperationsStatus{}}
	assert.False(t, s.isServerReady())
}

func Test_HandleServerOperationsNoFake(t *testing.T) {
	instance := &v1alpha1.Nexus{
		Spec:       v1alpha1.NexusSpec{},
		ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()},
	}
	cli := operatorclient.NewFakeClient(instance)

	status, err := HandleServerOperations(instance, cli)
	assert.NoError(t, err)
	assert.False(t, status.ServerReady)
}

func Test_handleServerOperations(t *testing.T) {
	instance := &v1alpha1.Nexus{
		Spec:       v1alpha1.NexusSpec{},
		ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()},
		Status: v1alpha1.NexusStatus{
			DeploymentStatus: appv1.DeploymentStatus{
				AvailableReplicas: 1,
			},
		},
	}
	svc := &corev1.Service{
		ObjectMeta: meta.DefaultObjectMeta(instance),
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:     "http",
					Protocol: corev1.ProtocolTCP,
					Port:     8081,
					TargetPort: intstr.IntOrString{
						IntVal: 8081,
					},
				},
			},
			Selector:        meta.GenerateLabels(instance),
			SessionAffinity: corev1.ServiceAffinityNone,
		},
	}
	cli := operatorclient.NewFakeClient(instance, svc, &corev1.Secret{ObjectMeta: v1.ObjectMeta{Name: instance.Name, Namespace: instance.Namespace}})
	status, err := handleServerOperations(instance, cli, nexusAPIFakeBuilder)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.CommunityRepositoriesCreated)
	assert.True(t, status.OperatorUserCreated)
	assert.True(t, status.ServerReady)
	// see: https://github.com/m88i/aicura/issues/18
	assert.False(t, status.MavenCentralUpdated)
}

func Test_handleServerOperationsNoEndpoint(t *testing.T) {
	instance := &v1alpha1.Nexus{
		Spec:       v1alpha1.NexusSpec{},
		ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()},
		Status: v1alpha1.NexusStatus{
			DeploymentStatus: appv1.DeploymentStatus{
				AvailableReplicas: 1,
			},
		},
	}
	cli := operatorclient.NewFakeClient(instance)
	status, err := handleServerOperations(instance, cli, nexusAPIFakeBuilder)
	assert.NoError(t, err)
	assert.NotNil(t, status)
	assert.False(t, status.CommunityRepositoriesCreated)
	assert.False(t, status.OperatorUserCreated)
	assert.False(t, status.ServerReady)
	assert.NotEmpty(t, status.Reason)
	// see: https://github.com/m88i/aicura/issues/18
	assert.False(t, status.MavenCentralUpdated)
}
