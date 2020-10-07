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

package validation

import (
	corev1 "k8s.io/api/core/v1"
	k8sres "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
)

const (
	NexusCommunityImage = "docker.io/sonatype/nexus3"
	NexusCertifiedImage = "registry.connect.redhat.com/sonatype/nexus-repository-manager"

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

	DefaultPersistence = v1alpha1.NexusPersistence{
		Persistent:   false,
		VolumeSize:   DefaultVolumeSize,
		StorageClass: "",
	}

	DefaultNetworking = v1alpha1.NexusNetworking{
		Expose: false,
		TLS:    DefaultTLS,
	}

	DefaultTLS = v1alpha1.NexusNetworkingTLS{
		Mandatory:  false,
		SecretName: "",
	}

	DefaultUpdate = v1alpha1.NexusAutomaticUpdate{
		// this isn't really the default, but we need this off for most tests anyway
		Disabled: true,
	}

	AllDefaultsCommunityNexus = v1alpha1.Nexus{
		ObjectMeta: metav1.ObjectMeta{Name: "default-community-nexus", Namespace: "default"},
		Spec: v1alpha1.NexusSpec{
			Replicas:                    0,
			Image:                       NexusCommunityImage,
			ImagePullPolicy:             "",
			AutomaticUpdate:             DefaultUpdate,
			Resources:                   DefaultResources,
			Persistence:                 DefaultPersistence,
			UseRedHatImage:              false,
			GenerateRandomAdminPassword: false,
			Networking:                  DefaultNetworking,
			ServiceAccountName:          "default-community-nexus",
			LivenessProbe:               DefaultProbe.DeepCopy(),
			ReadinessProbe:              DefaultProbe.DeepCopy(),
		},
	}
)
