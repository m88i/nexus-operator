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
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/meta"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const defaultVolumeSize = "10Gi"

func newPVC(nexus *v1alpha1.Nexus) *corev1.PersistentVolumeClaim {
	if len(nexus.Spec.Persistence.VolumeSize) == 0 {
		nexus.Spec.Persistence.VolumeSize = defaultVolumeSize
	}

	accessMode := corev1.ReadWriteOnce
	if nexus.Spec.Replicas > 1 {
		accessMode = corev1.ReadWriteMany
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: meta.DefaultObjectMeta(nexus),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				accessMode,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(nexus.Spec.Persistence.VolumeSize),
				},
			},
		},
	}

	if len(nexus.Spec.Persistence.StorageClass) > 0 {
		pvc.Spec.StorageClassName = &nexus.Spec.Persistence.StorageClass
	}

	return pvc
}
