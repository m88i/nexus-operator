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

package deployment

import (
	"fmt"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/resource/meta"
	"sort"
	"strconv"
	"strings"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	jvmArgsXms                 = "-Xms"
	jvmArgsXmx                 = "-Xmx"
	jvmArgsMaxMemSize          = "-XX:MaxDirectMemorySize"
	jvmArgsUserRoot            = "-Djava.util.prefs.userRoot"
	jvmArgRandomPassword       = "-Dnexus.security.randompassword"
	heapSizeDefault            = "1718m"
	maxDirectMemorySizeDefault = "2148m"
	nexusDataDir               = "/nexus-data"
	nexusContainerName         = "nexus-server"
)

var (
	jvmArgsMap = map[string]string{
		jvmArgsXms:           heapSizeDefault,
		jvmArgsXmx:           heapSizeDefault,
		jvmArgsMaxMemSize:    maxDirectMemorySizeDefault,
		jvmArgsUserRoot:      "${NEXUS_DATA}/javaprefs",
		jvmArgRandomPassword: "false",
	}

	nexusUID = int64(200)
)

func newDeployment(nexus *v1alpha1.Nexus) *appsv1.Deployment {
	deployment := &appsv1.Deployment{
		ObjectMeta: meta.DefaultObjectMeta(nexus),
		Spec: appsv1.DeploymentSpec{
			Replicas: &nexus.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: meta.GenerateLabels(nexus),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: meta.DefaultObjectMeta(nexus),
				Spec: corev1.PodSpec{
					ServiceAccountName: nexus.Spec.ServiceAccountName,
					Containers: []corev1.Container{
						{
							Name: nexusContainerName,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: NexusServicePort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources: nexus.Spec.Resources,
							Image:     nexus.Spec.Image,
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
	}

	addVolume(nexus, deployment)
	addProbes(nexus, deployment)
	applyJVMArgs(nexus, deployment)
	applySecurityContext(nexus, deployment)
	applyPullPolicy(nexus, deployment)

	return deployment
}

func applyPullPolicy(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	if len(nexus.Spec.ImagePullPolicy) > 0 {
		deployment.Spec.Template.Spec.Containers[0].ImagePullPolicy = nexus.Spec.ImagePullPolicy
	}
}

func addProbes(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	livenessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/service/rest/v1/status",
				Port: intstr.IntOrString{
					IntVal: NexusServicePort,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: nexus.Spec.LivenessProbe.InitialDelaySeconds,
		TimeoutSeconds:      nexus.Spec.LivenessProbe.TimeoutSeconds,
		FailureThreshold:    nexus.Spec.LivenessProbe.FailureThreshold,
		PeriodSeconds:       nexus.Spec.LivenessProbe.PeriodSeconds,
		SuccessThreshold:    nexus.Spec.LivenessProbe.SuccessThreshold,
	}

	readinessProbe := &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path: "/service/rest/v1/status",
				Port: intstr.IntOrString{
					IntVal: NexusServicePort,
				},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		InitialDelaySeconds: nexus.Spec.ReadinessProbe.InitialDelaySeconds,
		TimeoutSeconds:      nexus.Spec.ReadinessProbe.TimeoutSeconds,
		FailureThreshold:    nexus.Spec.ReadinessProbe.FailureThreshold,
		PeriodSeconds:       nexus.Spec.ReadinessProbe.PeriodSeconds,
		SuccessThreshold:    nexus.Spec.ReadinessProbe.SuccessThreshold,
	}

	deployment.Spec.Template.Spec.Containers[0].LivenessProbe = livenessProbe
	deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = readinessProbe
}

func addVolume(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	if nexus.Spec.Persistence.Persistent {
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: fmt.Sprintf("%s-data", nexus.Name),
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: nexus.Name,
						ReadOnly:  false,
					},
				},
			},
		}
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{
			{
				Name:      fmt.Sprintf("%s-data", nexus.Name),
				MountPath: nexusDataDir,
			},
		}
	}
}

func applyJVMArgs(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	jvmMemory, directMemSize := calculateJVMMemory(deployment.Spec.Template.Spec.Containers[0].Resources.Limits)
	jvmArgsMap[jvmArgsXms] = jvmMemory
	jvmArgsMap[jvmArgsXmx] = jvmMemory
	jvmArgsMap[jvmArgsMaxMemSize] = directMemSize
	jvmArgsMap[jvmArgRandomPassword] = strconv.FormatBool(nexus.Spec.GenerateRandomAdminPassword)

	// we cannot guarantee the key order when transforming the map into a single string.
	// this might create different strings across the reconciliation loop, causing the comparator to accuse the deployment to be different
	// that's why we have to sort the map every time.
	var jvmArgs strings.Builder
	var jvmArgsKeys []string
	for key := range jvmArgsMap {
		jvmArgsKeys = append(jvmArgsKeys, key)
	}
	sort.Strings(jvmArgsKeys)
	for _, key := range jvmArgsKeys {
		jvmArgs.WriteString(key)
		if key != jvmArgsXms && key != jvmArgsXmx {
			jvmArgs.WriteString("=")
		}
		jvmArgs.WriteString(jvmArgsMap[key])
		jvmArgs.WriteString(" ")
	}

	deployment.Spec.Template.Spec.Containers[0].Env =
		append(deployment.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name:  jvmArgsEnvKey,
				Value: jvmArgs.String(),
			})
}

func calculateJVMMemory(limits corev1.ResourceList) (jvmMemory, directMemSize string) {
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

func applySecurityContext(nexus *v1alpha1.Nexus, deployment *appsv1.Deployment) {
	var podSecContext *corev1.PodSecurityContext
	if !nexus.Spec.UseRedHatImage {
		podSecContext = &corev1.PodSecurityContext{FSGroup: &nexusUID, RunAsUser: &nexusUID, SupplementalGroups: []int64{nexusUID}}
	}
	deployment.Spec.Template.Spec.SecurityContext = podSecContext
}
