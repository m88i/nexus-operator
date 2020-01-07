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
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

// GetComparator will create the default comparator for the Nexus instance
// The comparator can be used to compare two different sets of resources and update them accordingly
func GetComparator() compare.MapComparator {
	resourceComparator := compare.DefaultComparator()

	pvcType := reflect.TypeOf(v1.PersistentVolumeClaim{})
	svcType := reflect.TypeOf(v1.Service{})
	routeType := reflect.TypeOf(routev1.Route{})
	deploymentType := reflect.TypeOf(appsv1.Deployment{})

	resourceComparator.SetComparator(pvcType, resourceComparator.GetDefaultComparator())
	resourceComparator.SetComparator(svcType, resourceComparator.GetComparator(svcType))
	resourceComparator.SetComparator(deploymentType, resourceComparator.GetComparator(deploymentType))
	resourceComparator.SetComparator(routeType, resourceComparator.GetComparator(routeType))

	return compare.MapComparator{Comparator: resourceComparator}
}
