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
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

var log = logf.Log.WithName("controller_nexus")

// GetDeployedResources will fetch for the resources managed by the Nexus instance deployed in the cluster
func GetDeployedResources(nexus *v1alpha1.Nexus, client client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	reader := read.New(client).WithNamespace(nexus.Namespace).WithOwnerObject(nexus)
	resources, err = reader.ListAll(&v1.PersistentVolumeClaimList{}, &v1.ServiceList{}, &appsv1.DeploymentList{})
	if err != nil {
		log.Error(err, "Failed to fetch deployed Nexus resources")
		return nil, err
	}
	if resources == nil {
		log.Info("No deployed resources found")
	}
	log.Info("Number of deployed ", "resources", len(resources))
	return resources, nil
}

// CreateRequiredResources will create the requests resources as it's supposed to be
func CreateRequiredResources(nexus *v1alpha1.Nexus) (resources map[reflect.Type][]resource.KubernetesResource) {
	logger := log.WithValues("Nexus.Namespace", nexus.Namespace, "Nexus.Name", nexus.Name)
	logger.Info("Creating resources structures")
	resources = make(map[reflect.Type][]resource.KubernetesResource)

	pvc := newPVC(nexus)
	if pvc != nil {
		resources[reflect.TypeOf(v1.PersistentVolumeClaim{})] = []resource.KubernetesResource{pvc}
	}
	resources[reflect.TypeOf(appsv1.Deployment{})] = []resource.KubernetesResource{newDeployment(nexus, pvc)}
	resources[reflect.TypeOf(v1.Service{})] = []resource.KubernetesResource{newService(nexus)}

	return resources
}
