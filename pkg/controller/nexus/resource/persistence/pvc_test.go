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

package persistence

import (
	"testing"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/validation"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_newPVC_defaultValues(t *testing.T) {
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: v1.ObjectMeta{
			Name:      appName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: true,
				VolumeSize: validation.DefaultVolumeSize,
			},
		},
	}
	pvc := newPVC(nexus)

	assert.Len(t, pvc.Spec.AccessModes, 1)
	assert.Equal(t, corev1.ReadWriteOnce, pvc.Spec.AccessModes[0])
	assert.Equal(t, resource.MustParse(validation.DefaultVolumeSize), pvc.Spec.Resources.Requests["storage"])
}

func Test_newPVC_moreThanOneReplica(t *testing.T) {
	appName := "nexus3"
	volumeSize := "20Gi"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: v1.ObjectMeta{
			Name:      appName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.NexusSpec{
			Replicas: 2,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: true,
				VolumeSize: volumeSize,
			},
		},
	}
	pvc := newPVC(nexus)

	assert.Len(t, pvc.Spec.AccessModes, 1)
	assert.Equal(t, corev1.ReadWriteMany, pvc.Spec.AccessModes[0])
	assert.Equal(t, resource.MustParse(volumeSize), pvc.Spec.Resources.Requests["storage"])
}
