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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/meta"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/validation"
)

func Test_newDeployment_WithoutPersistence(t *testing.T) {
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.NexusSpec{
			AutomaticUpdate: v1alpha1.NexusAutomaticUpdate{Disabled: true},
			Replicas:        1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: false,
			},
			// a valid Liveness Probe should have successThreshold == 1
			// but we don't care about that here (this is tested on the manager's tests)
			LivenessProbe:  validation.DefaultProbe,
			ReadinessProbe: validation.DefaultProbe,
			Image:          validation.NexusCommunityImage,
		},
	}
	deployment := newDeployment(nexus)

	assert.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, validation.NexusCommunityImage, deployment.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, int32(1), *deployment.Spec.Replicas)

	assert.Equal(t, int32(NexusServicePort), deployment.Spec.Template.Spec.Containers[0].LivenessProbe.HTTPGet.Port.IntVal)
	assert.Equal(t, int32(NexusServicePort), deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.HTTPGet.Port.IntVal)

	assert.Len(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, 0)
	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 0)

	assert.Equal(t, appName, deployment.Labels[meta.AppLabel])
	assert.Equal(t, appName, deployment.Spec.Template.Labels[meta.AppLabel])
	assert.Equal(t, appName, deployment.Spec.Selector.MatchLabels[meta.AppLabel])
}

func Test_newDeployment_WithPersistence(t *testing.T) {
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.NexusSpec{
			AutomaticUpdate: v1alpha1.NexusAutomaticUpdate{Disabled: true},
			Replicas:        1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: true,
			},
			// a valid Liveness Probe should have successThreshold == 1
			// but we don't care about that here (this is tested on the manager's tests)
			LivenessProbe:  validation.DefaultProbe,
			ReadinessProbe: validation.DefaultProbe,
		},
	}
	deployment := newDeployment(nexus)

	assert.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	assert.Len(t, deployment.Spec.Template.Spec.Containers[0].VolumeMounts, 1)
	assert.Len(t, deployment.Spec.Template.Spec.Volumes, 1)
	assert.Equal(t, nexusDataDir, deployment.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
}

// see: https://stackoverflow.com/questions/50804915/kubernetes-size-definitions-whats-the-difference-of-gi-and-g
func Test_calculateJVMMemory(t *testing.T) {
	type args struct {
		limits corev1.ResourceList
	}
	tests := []struct {
		name              string
		args              args
		wantJvmMemory     string
		wantDirectMemSize string
	}{
		{
			"2 Giga",
			args{limits: map[corev1.ResourceName]resource.Quantity{corev1.ResourceMemory: resource.MustParse("2G")}},
			"1600m",
			"2000m",
		},
		{
			"1.4 Mega",
			args{limits: map[corev1.ResourceName]resource.Quantity{corev1.ResourceMemory: resource.MustParse("1400M")}},
			"1120m",
			"1400m",
		},
		{
			"0.4 Mega",
			args{limits: map[corev1.ResourceName]resource.Quantity{corev1.ResourceMemory: resource.MustParse("400M")}},
			"320m",
			"400m",
		},
		{
			"10 Giga",
			args{limits: map[corev1.ResourceName]resource.Quantity{corev1.ResourceMemory: resource.MustParse("10G")}},
			"8000m",
			"10000m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotJvmMemory, gotDirectMemSize := calculateJVMMemory(tt.args.limits)
			if gotJvmMemory != tt.wantJvmMemory {
				t.Errorf("calculateJVMMemory() gotJvmMemory = %v, want %v", gotJvmMemory, tt.wantJvmMemory)
			}
			if gotDirectMemSize != tt.wantDirectMemSize {
				t.Errorf("calculateJVMMemory() gotDirectMemSize = %v, want %v", gotDirectMemSize, tt.wantDirectMemSize)
			}
		})
	}
}

func Test_customProbes(t *testing.T) {
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: true,
			},
			AutomaticUpdate: v1alpha1.NexusAutomaticUpdate{Disabled: true},
			LivenessProbe: &v1alpha1.NexusProbe{
				FailureThreshold:    1,
				PeriodSeconds:       10,
				InitialDelaySeconds: 0,
				SuccessThreshold:    3,
				TimeoutSeconds:      15,
			},
			ReadinessProbe: validation.DefaultProbe,
		},
	}
	deployment := newDeployment(nexus)

	assert.Len(t, deployment.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, int32(1), deployment.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold)
	assert.Equal(t, int32(0), deployment.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds)
	assert.Equal(t, validation.DefaultProbe.InitialDelaySeconds, deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.InitialDelaySeconds)
}

func Test_applyJVMArgs_withRandomPassword(t *testing.T) {
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: true,
			},
			AutomaticUpdate:             v1alpha1.NexusAutomaticUpdate{Disabled: true},
			GenerateRandomAdminPassword: true,
			// a valid Liveness Probe should have successThreshold == 1
			// but we don't care about that here (this is tested on the manager's tests)
			LivenessProbe:  validation.DefaultProbe,
			ReadinessProbe: validation.DefaultProbe,
		},
	}
	deployment := newDeployment(nexus)

	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env[0].Value, strings.Join([]string{jvmArgRandomPassword, "true"}, "="))
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env[0].Value, strings.Join([]string{jvmArgsXms, heapSizeDefault}, ""))
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env[0].Value, strings.Join([]string{jvmArgsXmx, heapSizeDefault}, ""))
}

func Test_applyJVMArgs_withDefaultValues(t *testing.T) {
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: true,
			},
			AutomaticUpdate: v1alpha1.NexusAutomaticUpdate{Disabled: true},
			// a valid Liveness Probe should have successThreshold == 1
			// but we don't care about that here (this is tested on the manager's tests)
			LivenessProbe:  validation.DefaultProbe,
			ReadinessProbe: validation.DefaultProbe,
		},
	}
	deployment := newDeployment(nexus)

	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env[0].Value, strings.Join([]string{jvmArgRandomPassword, "false"}, "="))
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env[0].Value, strings.Join([]string{jvmArgsXms, heapSizeDefault}, ""))
	assert.Contains(t, deployment.Spec.Template.Spec.Containers[0].Env[0].Value, strings.Join([]string{jvmArgsXmx, heapSizeDefault}, ""))
}
