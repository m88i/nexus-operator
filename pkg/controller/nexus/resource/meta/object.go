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

package meta

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

const AppLabel = "app"

func DefaultObjectMeta(nexus *v1alpha1.Nexus) v1.ObjectMeta {
	return v1.ObjectMeta{
		Namespace: nexus.Namespace,
		Name:      nexus.Name,
		Labels:    GenerateLabels(nexus),
	}
}

func GenerateLabels(nexus *v1alpha1.Nexus) map[string]string {
	nexusAppLabels := map[string]string{}
	nexusAppLabels[AppLabel] = nexus.Name
	return nexusAppLabels
}
