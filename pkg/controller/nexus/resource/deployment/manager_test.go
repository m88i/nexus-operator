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
			Resources:          DefaultResources,
			Image:              NexusCommunityLatestImage,
			LivenessProbe:      DefaultProbe.DeepCopy(),
			ReadinessProbe:     DefaultProbe.DeepCopy(),
		},
	}

	allDefaultsRedHatNexus = func() *v1alpha1.Nexus {
		nexus := allDefaultsCommunityNexus.DeepCopy()
		nexus.Spec.UseRedHatImage = true
		nexus.Spec.Image = NexusCertifiedLatestImage
		return nexus
	}()
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

func TestManager_setDefaults(t *testing.T) {
	minimumDefaultProbe := &v1alpha1.NexusProbe{
		InitialDelaySeconds: 0,
		TimeoutSeconds:      1,
		PeriodSeconds:       1,
		SuccessThreshold:    1,
		FailureThreshold:    1,
	}

	tests := []struct {
		name  string
		input *v1alpha1.Nexus
		want  *v1alpha1.Nexus
	}{
		{
			"'spec.serviceAccountName' left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ServiceAccountName = ""
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.resources' left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Resources = corev1.ResourceRequirements{}
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.useRedHatImage' set to true and 'spec.image' not left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.UseRedHatImage = true
				return nexus
			}(),
			allDefaultsRedHatNexus,
		}, {
			"'spec.useRedHatImage' set to false and 'spec.image' left blank",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.Image = ""
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.successThreshold' not equal to 1",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe.SuccessThreshold = 2
				return nexus
			}(),
			allDefaultsCommunityNexus,
		},
		{
			"'spec.livenessProbe.*' and 'spec.readinessProbe.*' don't meet minimum values",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = &v1alpha1.NexusProbe{
					InitialDelaySeconds: -1,
					TimeoutSeconds:      -1,
					PeriodSeconds:       -1,
					SuccessThreshold:    -1,
					FailureThreshold:    -1,
				}
				nexus.Spec.ReadinessProbe = nexus.Spec.LivenessProbe.DeepCopy()
				return nexus
			}(),
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = minimumDefaultProbe.DeepCopy()
				nexus.Spec.ReadinessProbe = minimumDefaultProbe.DeepCopy()
				return nexus
			}(),
		},
		{
			"Unset 'spec.livenessProbe' and 'spec.readinessProbe'",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.LivenessProbe = nil
				nexus.Spec.ReadinessProbe = nil
				return nexus
			}(),
			allDefaultsCommunityNexus.DeepCopy(),
		},
		{
			"Invalid 'spec.imagePullPolicy'",
			func() *v1alpha1.Nexus {
				nexus := allDefaultsCommunityNexus.DeepCopy()
				nexus.Spec.ImagePullPolicy = "invalid"
				return nexus
			}(),
			allDefaultsCommunityNexus.DeepCopy(),
		},
	}

	for _, tt := range tests {
		manager := &Manager{
			nexus: tt.input,
		}
		manager.setDefaults()
		if !reflect.DeepEqual(manager.nexus, tt.want) {
			t.Errorf("TestManager_setDefaults() - %s\nWant: %+v\tGot: %+v", tt.name, tt.want, *manager.nexus)
		}
	}
}

func TestManager_GetRequiredResources(t *testing.T) {
	// first, let's test with mgr that has not been init
	mgr := &Manager{}
	resources, err := mgr.GetDeployedResources()
	assert.Nil(t, resources)
	assert.EqualError(t, err, mgrNotInit)

	// correctness of the generated resources is tested elsewhere
	// here we just want to check if they have been created and returned
	mgr = &Manager{
		nexus:  allDefaultsCommunityNexus,
		client: test.NewFakeClientBuilder().Build(),
	}
	resources, err = mgr.GetRequiredResources()
	assert.Nil(t, err)
	// a deployment and a service are _always_ created, so both should always be present
	assert.Len(t, resources, 2)
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&corev1.Service{})))
	assert.True(t, test.ContainsType(resources, reflect.TypeOf(&appsv1.Deployment{})))
}

func TestManager_GetDeployedResources(t *testing.T) {
	// first, let's test with mgr that has not been init
	mgr := &Manager{}
	resources, err := mgr.GetDeployedResources()
	assert.Nil(t, resources)
	assert.EqualError(t, err, mgrNotInit)

	// now a valid mgr, but no deployed resources
	fakeClient := test.NewFakeClientBuilder().Build()
	mgr = &Manager{
		nexus:  allDefaultsCommunityNexus,
		client: fakeClient,
	}
	resources, err = mgr.GetDeployedResources()
	assert.Nil(t, resources)
	assert.Len(t, resources, 0)
	assert.NoError(t, err)

	// now a valid mgr with deployed resources
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
