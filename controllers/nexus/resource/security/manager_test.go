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

package security

import (
	ctx "context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/client"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/m88i/nexus-operator/pkg/framework/kind"
	"github.com/m88i/nexus-operator/pkg/logger"
	"github.com/m88i/nexus-operator/pkg/test"
)

var baseNexus = &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "nexus"}, Spec: v1alpha1.NexusSpec{ServiceAccountName: "nexus"}}

func TestNewManager(t *testing.T) {
	// default-setting logic is tested elsewhere
	// so here we just check if the resulting manager took in the arguments correctly
	nexus := baseNexus
	cli := client.NewFakeClient()
	want := &Manager{
		nexus:  nexus,
		client: cli,
	}
	got := NewManager(nexus, cli)
	assert.Equal(t, want.nexus, got.nexus)
	assert.Equal(t, want.client, got.client)
}

func TestManager_GetRequiredResources(t *testing.T) {
	// correctness of the generated resources is tested elsewhere
	// here we just want to check if they have been created and returned
	mgr := &Manager{
		nexus:  baseNexus.DeepCopy(),
		client: client.NewFakeClient(),
		log:    logger.GetLoggerWithResource("test", baseNexus),
	}

	// the default service accout is _always_ created
	// even if the user specified a different one
	resources, err := mgr.GetRequiredResources()
	assert.Nil(t, err)
	assert.Len(t, resources, 2)
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&corev1.ServiceAccount{})))
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&corev1.Secret{})))
}

func TestManager_GetDeployedResources(t *testing.T) {
	// first with no deployed resources
	fakeClient := client.NewFakeClient()
	mgr := &Manager{
		nexus:  baseNexus,
		client: fakeClient,
	}
	resources, err := mgr.GetDeployedResources()
	assert.Nil(t, resources)
	assert.Len(t, resources, 0)
	assert.NoError(t, err)

	// now with a deployed Service Account
	pvc := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: mgr.nexus.Name, Namespace: mgr.nexus.Namespace}}
	assert.NoError(t, mgr.client.Create(ctx.TODO(), pvc))

	resources, err = mgr.GetDeployedResources()
	assert.NoError(t, err)
	assert.Len(t, resources, 1)
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&corev1.ServiceAccount{})))

	// make the client return a mocked 500 response to test errors other than NotFound
	mockErrorMsg := "mock 500"
	fakeClient.SetMockErrorForOneRequest(errors.NewInternalError(fmt.Errorf(mockErrorMsg)))
	resources, err = mgr.GetDeployedResources()
	assert.Nil(t, resources)
	assert.Contains(t, err.Error(), mockErrorMsg)
}

func TestManager_getDeployedSvcAccnt(t *testing.T) {
	mgr := &Manager{
		nexus:  baseNexus,
		client: client.NewFakeClient(),
	}

	// first, test without creating the svcAccnt
	err := client.Fetch(mgr.client, framework.Key(mgr.nexus), managedObjectsRef[kind.SvcAccountKind], kind.SvcAccountKind)
	assert.True(t, errors.IsNotFound(err))

	// now test after creating the svcAccnt
	svcAccnt := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: mgr.nexus.Name, Namespace: mgr.nexus.Namespace}}
	assert.NoError(t, mgr.client.Create(ctx.TODO(), svcAccnt))
	err = client.Fetch(mgr.client, framework.Key(svcAccnt), svcAccnt, kind.SvcAccountKind)
	assert.NotNil(t, svcAccnt)
	assert.NoError(t, err)
}

func TestManager_GetCustomComparator(t *testing.T) {
	// the nexus and the client should have no effect on the
	// comparator functions offered by the manager
	mgr := &Manager{}

	// there is no custom comparator function for Service Accounts
	pvcComp := mgr.GetCustomComparator(reflect.TypeOf(&corev1.ServiceAccount{}))
	assert.Nil(t, pvcComp)

	// there is no custom comparator function for Service Accounts
	secretComp := mgr.GetCustomComparator(reflect.TypeOf(&corev1.Secret{}))
	assert.NotNil(t, secretComp)
}

func TestManager_GetCustomComparators(t *testing.T) {
	// the nexus and the client should have no effect on the
	// comparator functions offered by the manager
	mgr := &Manager{}

	// there is no custom comparator function for Service Accounts
	comparators := mgr.GetCustomComparators()
	assert.Len(t, comparators, 1)
}
