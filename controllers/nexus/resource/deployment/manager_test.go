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

package deployment

import (
	ctx "context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
	"github.com/m88i/nexus-operator/pkg/test"
)

var (
	allDefaultsCommunityNexus = &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{Name: "nexus-test"},
		Spec: v1alpha1.NexusSpec{
			AutomaticUpdate:    v1alpha1.NexusAutomaticUpdate{Disabled: true},
			ServiceAccountName: "nexus-test",
			Resources:          v1alpha1.DefaultResources,
			Image:              v1alpha1.NexusCommunityImage,
			LivenessProbe:      v1alpha1.DefaultProbe.DeepCopy(),
			ReadinessProbe:     v1alpha1.DefaultProbe.DeepCopy(),
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
	got := NewManager(nexus, client)
	assert.Equal(t, want.nexus, got.nexus)
	assert.Equal(t, want.client, got.client)
}

func TestManager_GetRequiredResources(t *testing.T) {
	// correctness of the generated resources is tested elsewhere
	// here we just want to check if they have been created and returned
	mgr := &Manager{
		nexus:  allDefaultsCommunityNexus,
		client: test.NewFakeClientBuilder().Build(),
		log:    logger.GetLoggerWithResource("test", allDefaultsCommunityNexus),
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

func TestManager_GetCustomComparator(t *testing.T) {
	// the nexus and the client should have no effect on the
	// comparator functions offered by the manager
	mgr := &Manager{}

	// there is a custom comparator function for deployments, but not services
	deploymentComp := mgr.GetCustomComparator(reflect.TypeOf(&appsv1.Deployment{}))
	assert.NotNil(t, deploymentComp)
	svcComp := mgr.GetCustomComparator(reflect.TypeOf(&corev1.Service{}))
	assert.Nil(t, svcComp)
}

func TestManager_GetCustomComparators(t *testing.T) {
	// the nexus and the client should have no effect on the
	// comparator functions offered by the manager
	mgr := &Manager{}

	// there is a custom comparator for deployments
	comparators := mgr.GetCustomComparators()
	assert.Len(t, comparators, 1)
}

func Test_deploymentEqual(t *testing.T) {
	baseDeployment := newDeployment(allDefaultsCommunityNexus)
	baseDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
	tests := []struct {
		name      string
		req       *appsv1.Deployment
		dep       *appsv1.Deployment
		wantEqual bool
	}{
		{
			"Completely equal deployments",
			baseDeployment.DeepCopy(),
			baseDeployment.DeepCopy(),
			true,
		},
		{
			"Different name",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Name = "different"
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different namespace",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Namespace = "different"
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different labels",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Labels["different key"] = "different value"
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different replicas",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				r := int32(10)
				d.Spec.Replicas = &r
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different selector",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Selector = &metav1.LabelSelector{}
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different template object meta",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Name = "different"
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different volumes",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.Volumes = []corev1.Volume{}
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different service account name",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.ServiceAccountName = "different"
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different pod security context",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.SecurityContext = &corev1.PodSecurityContext{}
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different container name",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.Containers[0].Name = "different"
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different container ports",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{}
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different container resources",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.Containers[0].Resources = corev1.ResourceRequirements{}
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different container image",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.Containers[0].Image = "different"
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different container liveness probe",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.Containers[0].LivenessProbe = &corev1.Probe{}
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different container readiness probe",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.Containers[0].ReadinessProbe = &corev1.Probe{}
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different container env",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Template.Spec.Containers[0].Env = []corev1.EnvVar{}
				return d
			}(),
			baseDeployment.DeepCopy(),
			false,
		},
		{
			"Different field we don't care about (deployment strategy)",
			func() *appsv1.Deployment {
				d := baseDeployment.DeepCopy()
				d.Spec.Strategy = appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType}
				return d
			}(),
			baseDeployment.DeepCopy(),
			true,
		},
	}

	for _, tt := range tests {
		gotEqual := deploymentEqual(tt.dep, tt.req)
		if gotEqual != tt.wantEqual {
			t.Errorf("%s - wantEqual: %v\tgotEqual: %v", tt.name, tt.wantEqual, gotEqual)
		}
	}
}

func Test_equalPullPolicies(t *testing.T) {
	depDeployment := newDeployment(allDefaultsCommunityNexus)
	reqDeployment := depDeployment.DeepCopy()

	// first let's test an unspecified pull policy with no tag
	depDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
	assert.True(t, equalPullPolicies(depDeployment, reqDeployment))
	depDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullIfNotPresent
	assert.False(t, equalPullPolicies(depDeployment, reqDeployment))

	// now let's set the latest tag on the images so we can test for the PullAlways pull policy in that scenario as well
	depDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
	reqDeployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", v1alpha1.NexusCommunityImage, "latest")
	assert.True(t, equalPullPolicies(depDeployment, reqDeployment))

	// now with an actual tag and empty pullPolicy on the required deployment
	reqDeployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", v1alpha1.NexusCommunityImage, "3.25.0")
	depDeployment.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", v1alpha1.NexusCommunityImage, "3.25.0")
	depDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullIfNotPresent
	assert.True(t, equalPullPolicies(depDeployment, reqDeployment))

	// with the same pull policies
	reqDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullIfNotPresent
	assert.True(t, equalPullPolicies(depDeployment, reqDeployment))

	// with different pull policies
	depDeployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = corev1.PullAlways
	assert.False(t, equalPullPolicies(depDeployment, reqDeployment))
}
