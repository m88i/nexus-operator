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

package openshift

import (
	"k8s.io/client-go/discovery"
	"strings"
)

const (
	openshiftGroup = "openshift.io"
)

// IsOpenShift verify if the operator is running on OpenShift
func IsOpenShift(discovery discovery.DiscoveryInterface) (bool, error) {
	return hasGroup(openshiftGroup, discovery)
}

// hasGroup check if the given group name is available in the cluster
func hasGroup(group string, discovery discovery.DiscoveryInterface) (bool, error) {
	if discovery != nil {
		groups, err := discovery.ServerGroups()
		if err != nil {
			return false, err
		}
		for _, g := range groups.Groups {
			if strings.Contains(g.Name, group) {
				return true, nil
			}
		}
	}
	return false, nil
}
