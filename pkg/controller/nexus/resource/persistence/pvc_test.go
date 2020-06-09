//     Copyright 2019 Nexus Operator and/or its authors
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

package persistence

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
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
				VolumeSize: defaultVolumeSize,
			},
		},
	}
	pvc := newPVC(nexus)

	assert.Len(t, pvc.Spec.AccessModes, 1)
	assert.Equal(t, corev1.ReadWriteOnce, pvc.Spec.AccessModes[0])
	assert.Equal(t, resource.MustParse(defaultVolumeSize), pvc.Spec.Resources.Requests["storage"])
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
