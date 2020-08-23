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

package update

import (
	ctx "context"
	"fmt"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestMonitorUpdate(t *testing.T) {
	image := "image"
	baseDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
		},
	}
	nexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: "test"},
		Status:     v1alpha1.NexusStatus{},
		Spec:       v1alpha1.NexusSpec{Image: image},
	}
	c := test.NewFakeClientBuilder(nexus).Build()

	// Not in an update and will not start one
	deployedDep := baseDeployment.DeepCopy()
	requiredDep := baseDeployment.DeepCopy()
	deployedDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3.25.0")
	requiredDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3.25.0")

	err := HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.Nil(t, err)
	assert.Len(t, nexus.Status.UpdateConditions, 0)

	// Not in an update and will start one
	requiredDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3.25.1")

	err = HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.Nil(t, err)
	assert.Len(t, nexus.Status.UpdateConditions, 1)
	assert.Equal(t, fmt.Sprintf(updateStartFormat, "3.25.0", "3.25.1"), nexus.Status.UpdateConditions[0])

	// In an update and receives a new update
	deployedDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3.25.0")
	requiredDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3.25.2")

	err = HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.Nil(t, err)
	assert.Len(t, nexus.Status.UpdateConditions, 1)
	assert.Equal(t, fmt.Sprintf(updateStartFormat, "3.25.0", "3.25.2"), nexus.Status.UpdateConditions[0])

	// In an update and it's still progressing
	deployedDep.Spec.Template.Spec.Containers[0].Image = requiredDep.Spec.Template.Spec.Containers[0].Image
	err = HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.Nil(t, err)
	assert.Len(t, nexus.Status.UpdateConditions, 1)
	assert.Equal(t, fmt.Sprintf(updateStartFormat, "3.25.0", "3.25.2"), nexus.Status.UpdateConditions[0])

	// In an update and it succeeds
	deployedDep.Status.Conditions = []appsv1.DeploymentCondition{{
		Type:   appsv1.DeploymentProgressing,
		Reason: "NewReplicaSetAvailable",
	}}

	err = HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.Nil(t, err)
	assert.Len(t, nexus.Status.UpdateConditions, 2)
	assert.Equal(t, fmt.Sprintf(updateStartFormat, "3.25.0", "3.25.2"), nexus.Status.UpdateConditions[0])
	assert.Equal(t, fmt.Sprintf(updateOKFormat, "3.25.0", "3.25.2"), nexus.Status.UpdateConditions[1])
	assert.True(t, eventExists(c, successfulUpdateReason))

	// In an update and it fails
	nexus.Status.UpdateConditions = []string{fmt.Sprintf(updateStartFormat, "3.25.0", "3.25.2")}
	deployedDep.Status.Conditions = []appsv1.DeploymentCondition{{
		Type:   appsv1.DeploymentProgressing,
		Status: "False",
	}}

	err = HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.Nil(t, err)
	assert.Len(t, nexus.Status.UpdateConditions, 2)
	assert.Equal(t, fmt.Sprintf(updateStartFormat, "3.25.0", "3.25.2"), nexus.Status.UpdateConditions[0])
	assert.Equal(t, fmt.Sprintf(updateFailedFormat, "3.25.0", "3.25.2"), nexus.Status.UpdateConditions[1])
	assert.True(t, nexus.Spec.AutomaticUpdate.Disabled)
	assert.Equal(t, fmt.Sprintf("%s:%s", image, "3.25.0"), nexus.Spec.Image)
	assert.True(t, eventExists(c, failedUpdateReason))

	// In an update, it fails and rolling back fails
	nexus.Status.UpdateConditions = []string{fmt.Sprintf(updateStartFormat, "3.25.0", "3.25.2")}
	nexus.Spec.AutomaticUpdate.Disabled = false
	deployedDep.Status.Conditions = []appsv1.DeploymentCondition{{
		Type:   appsv1.DeploymentProgressing,
		Status: "False",
	}}
	c.SetMockError(fmt.Errorf("mock error"))

	err = HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.NotNil(t, err)
	assert.False(t, eventExists(c, failedUpdateReason))

	// In an update, but the conditions have been tempered with and it can't be parsed
	nexus.Status.UpdateConditions = []string{fmt.Sprintf(updateStartPrefix+"wrong format %s %s", "3.25.0", "3.25.2")}
	nexus.Spec.AutomaticUpdate.Disabled = false

	err = HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.Nil(t, err)
	assert.Nil(t, nexus.Status.UpdateConditions)

	// automatic updates are disabled
	nexus.Spec.AutomaticUpdate.Disabled = true

	err = HandleUpdate(nexus, deployedDep, requiredDep, c.Scheme(), c)
	assert.Nil(t, err)
	assert.Nil(t, nexus.Status.UpdateConditions)
}

func eventExists(c client.Client, reason string) bool {
	eventList := &corev1.EventList{}
	_ = c.List(ctx.Background(), eventList)
	for _, event := range eventList.Items {
		if event.Reason == reason {
			return true
		}
	}
	return false
}

func Test_alreadyUpdating(t *testing.T) {
	// We already tested most behaviors in TestMonitorUpdate
	// There was an update, but it's done now
	previousTag := "3.25.0"
	targetTag := "3.25.1"
	nexus := &v1alpha1.Nexus{Status: v1alpha1.NexusStatus{}}
	nexus.Status.UpdateConditions = []string{fmt.Sprintf(updateStartFormat, previousTag, targetTag), fmt.Sprintf(updateOKFormat, previousTag, targetTag)}
	assert.False(t, alreadyUpdating(nexus))
}

func Test_isNewUpdate(t *testing.T) {
	// We already tested most behaviors in TestMonitorUpdate
	image := "image"
	baseDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
		},
	}

	deployedDep := baseDeployment.DeepCopy()
	requiredDep := baseDeployment.DeepCopy()

	// invalid tag
	deployedDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3.25.0")
	requiredDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3..0")
	updating, _, _ := isNewUpdate(deployedDep, requiredDep)
	assert.False(t, updating)
}

func Test_differentImagesOrMinors(t *testing.T) {
	image := "image"
	baseDeployment := &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{}},
				},
			},
		},
	}

	deployedDep := baseDeployment.DeepCopy()
	requiredDep := baseDeployment.DeepCopy()

	// different images
	deployedDep.Spec.Template.Spec.Containers[0].Image = image
	assert.True(t, differentImagesOrMinors(deployedDep, requiredDep))

	// deployed using no tag (same as 'latest')
	requiredDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3.25.0")
	assert.True(t, differentImagesOrMinors(deployedDep, requiredDep))

	// invalid deployed tag
	deployedDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3..0")
	assert.True(t, differentImagesOrMinors(deployedDep, requiredDep))

	// same minor
	deployedDep.Spec.Template.Spec.Containers[0].Image = fmt.Sprintf("%s:%s", image, "3.25.0")
	assert.False(t, differentImagesOrMinors(deployedDep, requiredDep))
}
