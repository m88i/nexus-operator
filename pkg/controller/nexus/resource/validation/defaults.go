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

package validation

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	k8sres "k8s.io/apimachinery/pkg/api/resource"
)

const (
	NexusCommunityLatestImage = "docker.io/sonatype/nexus3:latest"
	NexusCertifiedLatestImage = "registry.connect.redhat.com/sonatype/nexus-repository-manager"

	DefaultVolumeSize = "10Gi"

	probeDefaultInitialDelaySeconds = int32(240)
	probeDefaultTimeoutSeconds      = int32(15)
	probeDefaultPeriodSeconds       = int32(10)
	probeDefaultSuccessThreshold    = int32(1)
	probeDefaultFailureThreshold    = int32(3)
)

var (
	DefaultResources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    k8sres.MustParse("2"),
			corev1.ResourceMemory: k8sres.MustParse("2Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    k8sres.MustParse("1"),
			corev1.ResourceMemory: k8sres.MustParse("2Gi"),
		},
	}

	DefaultProbe = &v1alpha1.NexusProbe{
		InitialDelaySeconds: probeDefaultInitialDelaySeconds,
		TimeoutSeconds:      probeDefaultTimeoutSeconds,
		PeriodSeconds:       probeDefaultPeriodSeconds,
		SuccessThreshold:    probeDefaultSuccessThreshold,
		FailureThreshold:    probeDefaultFailureThreshold,
	}
)
