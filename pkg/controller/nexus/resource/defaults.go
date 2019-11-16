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
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// See: https://hub.docker.com/r/sonatype/nexus3/
const (
	appLabel                  = "app"
	nexusServicePort          = 8081
	nexusDataDir              = "/nexus-data"
	nexusCommunityLatestImage = "docker.io/sonatype/nexus3:latest"
	nexusCertifiedLatestImage = "registry.connect.redhat.com/sonatype/nexus-repository-manager"
	nexusVolumeSize           = "10Gi"
	nexusContainerName        = "nexus-server"
)

var (
	nexusPodReq = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("1"),
			corev1.ResourceMemory: resource.MustParse("2Gi"),
		},
	}
)

func applyLabels(nexus *v1alpha1.Nexus, objectMeta *v1.ObjectMeta) {
	objectMeta.Labels = generateLabels(nexus)
}

func generateLabels(nexus *v1alpha1.Nexus) map[string]string {
	nexusAppLabels := map[string]string{}
	nexusAppLabels[appLabel] = nexus.Name
	return nexusAppLabels
}
