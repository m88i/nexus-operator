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

package resource

import (
	"fmt"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	jvmArgsEnvKey = "INSTALL4J_ADD_VM_PARAMS"
	/*
		1. Xms
		2. Xmx
		3. MaxDirectMemorySize
	*/
	jvmArgsEnvValueFormat      = "-Xms%s -Xmx%s -XX:MaxDirectMemorySize=%s -Djava.util.prefs.userRoot=${NEXUS_DATA}/javaprefs"
	heapSizeDefault            = "1200m"
	maxDirectMemorySizeDefault = "2g"
	probeInitialDelaySeconds   = int32(120)
	probeTimeoutSeconds        = int32(15)
)

func newDeployment(nexus *v1alpha1.Nexus, pvc *v1.PersistentVolumeClaim) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nexus.Name,
			Namespace: nexus.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &nexus.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: generateLabels(nexus),
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					SecurityContext: &v1.PodSecurityContext{FSGroup: &nexusUID, RunAsUser: &nexusUID, SupplementalGroups: []int64{nexusUID}},
					Containers: []v1.Container{
						{
							Name: nexusContainerName,
							Ports: []v1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: nexusServicePort,
									Protocol:      v1.ProtocolTCP,
								},
							},
							ImagePullPolicy: v1.PullAlways,
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}

	applyLabels(nexus, &deployment.ObjectMeta)
	applyLabels(nexus, &deployment.Spec.Template.ObjectMeta)

	applyDefaultImage(nexus, deployment)
	applyDefaultResourceReqs(nexus, deployment)
	addVolume(nexus, pvc, deployment)
	addProbes(nexus, deployment)
	applyJVMArgs(deployment)
	addServiceAccount(nexus, deployment)

	return deployment
}

func applyDefaultImage(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	if nexus.Spec.UseRedHatImage {
		nexus.Spec.Image = nexusCertifiedLatestImage
	} else if len(nexus.Spec.Image) == 0 {
		nexus.Spec.UseRedHatImage = false
		nexus.Spec.Image = nexusCommunityLatestImage
	}

	deployment.Spec.Template.Spec.Containers[0].Image = nexus.Spec.Image
}

func applyDefaultResourceReqs(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	if &nexus.Spec.Resources == nil {
		nexus.Spec.Resources = nexusPodReq
	}
	deployment.Spec.Template.Spec.Containers[0].Resources = nexus.Spec.Resources
}

func addProbes(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	defaultProbe := &v1.Probe{
		Handler: v1.Handler{
			HTTPGet: &v1.HTTPGetAction{
				Path: "/",
				Port: intstr.IntOrString{
					IntVal: nexusServicePort,
				},
				Scheme: v1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: probeInitialDelaySeconds,
		TimeoutSeconds:      probeTimeoutSeconds,
	}

	deployment.Spec.Template.Spec.Containers[0].LivenessProbe = defaultProbe.DeepCopy()
	deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = defaultProbe.DeepCopy()

	if nexus.Spec.LivenessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe.FailureThreshold =
			util.EnsureMinimum(nexus.Spec.LivenessProbe.FailureThreshold, 1)
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe.InitialDelaySeconds =
			util.EnsureMinimum(nexus.Spec.LivenessProbe.InitialDelaySeconds, 0)
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe.PeriodSeconds =
			util.EnsureMinimum(nexus.Spec.LivenessProbe.PeriodSeconds, 1)
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe.SuccessThreshold =
			util.EnsureMinimum(nexus.Spec.LivenessProbe.SuccessThreshold, 1)
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe.TimeoutSeconds =
			util.EnsureMinimum(nexus.Spec.LivenessProbe.TimeoutSeconds, 1)
	}

	if nexus.Spec.ReadinessProbe != nil {
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.FailureThreshold =
			util.EnsureMinimum(nexus.Spec.ReadinessProbe.FailureThreshold, 1)
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.InitialDelaySeconds =
			util.EnsureMinimum(nexus.Spec.ReadinessProbe.InitialDelaySeconds, 0)
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.PeriodSeconds =
			util.EnsureMinimum(nexus.Spec.ReadinessProbe.PeriodSeconds, 1)
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.SuccessThreshold =
			util.EnsureMinimum(nexus.Spec.ReadinessProbe.SuccessThreshold, 1)
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe.TimeoutSeconds =
			util.EnsureMinimum(nexus.Spec.ReadinessProbe.TimeoutSeconds, 1)
	}
}

func addVolume(nexus *v1alpha1.Nexus, pvc *v1.PersistentVolumeClaim, deployment *appsv1.Deployment) {
	if nexus.Spec.Persistence.Persistent {
		deployment.Spec.Template.Spec.Volumes = []v1.Volume{
			{
				Name: fmt.Sprintf("%s-data", nexus.Name),
				VolumeSource: v1.VolumeSource{
					PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{
						ClaimName: pvc.Name,
						ReadOnly:  false,
					},
				},
			},
		}
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
			{
				Name:      fmt.Sprintf("%s-data", nexus.Name),
				MountPath: nexusDataDir,
			},
		}
	}
}

func applyJVMArgs(deployment *appsv1.Deployment) {
	jvmMemory, directMemSize := calculateJVMMemory(deployment.Spec.Template.Spec.Containers[0].Resources.Limits)
	deployment.Spec.Template.Spec.Containers[0].Env =
		append(deployment.Spec.Template.Spec.Containers[0].Env,
			v1.EnvVar{
				Name:  jvmArgsEnvKey,
				Value: fmt.Sprintf(jvmArgsEnvValueFormat, jvmMemory, jvmMemory, directMemSize),
			})
}

func calculateJVMMemory(limits v1.ResourceList) (jvmMemory, directMemSize string) {
	if limits != nil {
		memoryLimit := limits.Memory()
		if memoryLimit != nil {
			limitValue := memoryLimit.ScaledValue(resource.Mega)
			jvmMemory = fmt.Sprintf("%.0fm", float64(limitValue)*0.8)
			directMemSize = fmt.Sprintf("%dm", limitValue)
			return
		}
	}
	jvmMemory = heapSizeDefault
	directMemSize = maxDirectMemorySizeDefault
	return
}

func addServiceAccount(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	if len(nexus.Spec.ServiceAccountName) > 0 {
		deployment.Spec.Template.Spec.ServiceAccountName = nexus.Spec.ServiceAccountName
	} else {
		deployment.Spec.Template.Spec.ServiceAccountName = nexus.Name
	}
}
