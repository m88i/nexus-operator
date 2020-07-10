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

package security

import (
	ctx "context"
	"fmt"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

var baseNexus = &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "nexus"}, Spec: v1alpha1.NexusSpec{ServiceAccountName: "nexus"}}

func TestNewManager(t *testing.T) {
	// default-setting logic is tested elsewhere
	// so here we just check if the resulting manager took in the arguments correctly
	nexus := baseNexus
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

func TestManager_setDefaults(t *testing.T) {
	nexusName := "nexus"
	saName := "my-custom-sa"

	tests := []struct {
		name  string
		input *v1alpha1.Nexus
		want  *v1alpha1.Nexus
	}{
		{
			"unset 'spec.serviceAccountName",
			&v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: nexusName}, Spec: v1alpha1.NexusSpec{}},
			&v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: nexusName}, Spec: v1alpha1.NexusSpec{ServiceAccountName: nexusName}},
		},
		{

			"unset 'spec.serviceAccountName",
			&v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: nexusName}, Spec: v1alpha1.NexusSpec{ServiceAccountName: saName}},
			&v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: nexusName}, Spec: v1alpha1.NexusSpec{ServiceAccountName: saName}},
		},
	}

	for _, tt := range tests {
		manager := &Manager{
			nexus: tt.input,
		}
		manager.setDefaults()
		if !reflect.DeepEqual(manager.nexus, tt.want) {
			t.Errorf("TestManager_setDefaults() - %s\nWant: %v\tGot: %v", tt.name, tt.want, *manager.nexus)
		}
	}
}

func TestManager_GetRequiredResources(t *testing.T) {
	// correctness of the generated resources is tested elsewhere
	// here we just want to check if they have been created and returned
	mgr := &Manager{
		nexus:  baseNexus.DeepCopy(),
		client: test.NewFakeClientBuilder().Build(),
	}

	// the default service accout is _always_ created
	// even if the user specified a different one
	resources, err := mgr.GetRequiredResources()
	assert.Nil(t, err)
	assert.Len(t, resources, 1)
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&corev1.ServiceAccount{})))
}

func TestManager_GetDeployedResources(t *testing.T) {
	// first with no deployed resources
	fakeClient := test.NewFakeClientBuilder().Build()
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
		client: test.NewFakeClientBuilder().Build(),
	}

	// first, test without creating the svcAccnt
	svcAccnt, err := mgr.getDeployedSvcAccnt()
	assert.Nil(t, svcAccnt)
	assert.True(t, errors.IsNotFound(err))

	// now test after creating the svcAccnt
	svcAccnt = &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: mgr.nexus.Name, Namespace: mgr.nexus.Namespace}}
	assert.NoError(t, mgr.client.Create(ctx.TODO(), svcAccnt))
	svcAccnt, err = mgr.getDeployedSvcAccnt()
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
}

func TestManager_GetCustomComparators(t *testing.T) {
	// the nexus and the client should have no effect on the
	// comparator functions offered by the manager
	mgr := &Manager{}

	// there is no custom comparator function for Service Accounts
	comparators := mgr.GetCustomComparators()
	assert.Nil(t, comparators)
}
