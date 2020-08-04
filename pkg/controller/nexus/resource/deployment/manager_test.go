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

package deployment

import (
	ctx "context"
	"fmt"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/validation"
	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

var (
	allDefaultsCommunityNexus = &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{Name: "nexus-test"},
		Spec: v1alpha1.NexusSpec{
			ServiceAccountName: "nexus-test",
			Resources:          validation.DefaultResources,
			Image:              validation.NexusCommunityLatestImage,
			LivenessProbe:      validation.DefaultProbe.DeepCopy(),
			ReadinessProbe:     validation.DefaultProbe.DeepCopy(),
		},
	}
)

func TestNewManager(t *testing.T) {
	// default-setting logic is tested elsewhere
	// so here we just check if the resulting manager took in the arguments correctly
	nexus := allDefaultsCommunityNexus
	client := test.NewFakeClientBuilder().Build()
	want := &Manager{
		nexus:  nexus,
		client: client,
	}
	got := NewManager(*nexus, client)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("TestNewManager()\nWant: %+v\tGot: %+v", want, got)
	}
}

func TestManager_GetRequiredResources(t *testing.T) {
	// correctness of the generated resources is tested elsewhere
	// here we just want to check if they have been created and returned
	mgr := &Manager{
		nexus:  allDefaultsCommunityNexus,
		client: test.NewFakeClientBuilder().Build(),
	}
	resources, err := mgr.GetRequiredResources()
	assert.Nil(t, err)
	// a deployment and a service are _always_ created, so both should always be present
	assert.Len(t, resources, 2)
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&corev1.Service{})))
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&appsv1.Deployment{})))
}

func TestManager_GetDeployedResources(t *testing.T) {
	// first no deployed resources
	fakeClient := test.NewFakeClientBuilder().Build()
	mgr := &Manager{
		nexus:  allDefaultsCommunityNexus,
		client: fakeClient,
	}
	resources, err := mgr.GetDeployedResources()
	assert.Nil(t, resources)
	assert.Len(t, resources, 0)
	assert.NoError(t, err)

	// now with deployed resources
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: mgr.nexus.Name, Namespace: mgr.nexus.Namespace}}
	assert.NoError(t, mgr.client.Create(ctx.TODO(), deployment))

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: mgr.nexus.Name, Namespace: mgr.nexus.Namespace}}
	err = mgr.client.Create(ctx.TODO(), svc)
	assert.NoError(t, err)

	resources, err = mgr.GetDeployedResources()
	assert.Nil(t, err)
	// a deployment and a service are _always_ deployed, so both should always be present
	assert.Len(t, resources, 2)
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&corev1.Service{})))
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&appsv1.Deployment{})))

	// make the client return a mocked 500 response to test errors other than NotFound
	mockErrorMsg := "mock 500"
	fakeClient.SetMockErrorForOneRequest(errors.NewInternalError(fmt.Errorf(mockErrorMsg)))
	resources, err = mgr.GetDeployedResources()
	assert.Nil(t, resources)
	assert.Contains(t, err.Error(), mockErrorMsg)
}

func TestManager_getDeployedDeployment(t *testing.T) {
	mgr := &Manager{
		nexus:  allDefaultsCommunityNexus,
		client: test.NewFakeClientBuilder().Build(),
	}

	// first, test without creating the deployment
	deployment, err := mgr.getDeployedDeployment()
	assert.Nil(t, deployment)
	assert.True(t, errors.IsNotFound(err))

	// now test after creating the deployment
	deployment = &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: mgr.nexus.Name, Namespace: mgr.nexus.Namespace}}
	assert.NoError(t, mgr.client.Create(ctx.TODO(), deployment))
	deployment, err = mgr.getDeployedDeployment()
	assert.NotNil(t, deployment)
	assert.NoError(t, err)
}

func TestManager_getDeployedService(t *testing.T) {
	mgr := &Manager{
		nexus:  allDefaultsCommunityNexus,
		client: test.NewFakeClientBuilder().Build(),
	}

	// first, test without creating the service
	svc, err := mgr.getDeployedService()
	assert.Nil(t, svc)
	assert.True(t, errors.IsNotFound(err))

	// now test after creating the service
	svc = &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: mgr.nexus.Name, Namespace: mgr.nexus.Namespace}}
	assert.NoError(t, mgr.client.Create(ctx.TODO(), svc))
	svc, err = mgr.getDeployedService()
	assert.NotNil(t, svc)
	assert.NoError(t, err)
}

func TestManager_GetCustomComparator(t *testing.T) {
	// the nexus and the client should have no effect on the
	// comparator functions offered by the manager
	mgr := &Manager{}

	// there is no custom comparator function for services or deployments
	deploymentComp := mgr.GetCustomComparator(reflect.TypeOf(&appsv1.Deployment{}))
	assert.Nil(t, deploymentComp)
	svcComp := mgr.GetCustomComparator(reflect.TypeOf(&corev1.Service{}))
	assert.Nil(t, svcComp)
}

func TestManager_GetCustomComparators(t *testing.T) {
	// the nexus and the client should have no effect on the
	// comparator functions offered by the manager
	mgr := &Manager{}

	// there is no custom comparator function for services or deployments
	comparators := mgr.GetCustomComparators()
	assert.Nil(t, comparators)
}
