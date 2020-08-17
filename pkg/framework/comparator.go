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

package framework

import "github.com/RHsyseng/operator-utils/pkg/resource"

// AlwaysTrueComparator returns a comparator that always return 'true'.
// can be used on use cases where we want to create if not exists, but not update
func AlwaysTrueComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		return true
	}
}
