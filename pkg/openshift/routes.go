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
	"context"
	"fmt"
	v1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

const (
	routesGroup       = "route.openshift.io"
	httpPrefixSchema  = "http://"
	httpsPrefixSchema = "https://"
)

// IsRouteAvailable verifies if the current cluster has the Route API from OpenShift available
func IsRouteAvailable(discovery discovery.DiscoveryInterface) (bool, error) {
	if discovery != nil {
		groups, err := discovery.ServerGroups()
		if err != nil {
			return false, err
		}
		for _, group := range groups.Groups {
			if strings.Contains(group.Name, routesGroup) {
				return true, nil
			}
		}
	}
	return false, nil
}

// GetRouteURI will discover the route scheme based on the given namespaced name for the route
func GetRouteURI(client client.Client, discovery discovery.DiscoveryInterface, routeName types.NamespacedName) (uri string, err error) {
	ok, err := IsRouteAvailable(discovery)
	if err != nil {
		return "", err
	}
	if ok {
		route := &v1.Route{}
		if err := client.Get(context.TODO(), routeName, route); err != nil {
			if !errors.IsNotFound(err) {
				return "", err
			}
		}
		if len(route.Spec.Host) > 0 {
			if nil == route.Spec.TLS {
				return fmt.Sprintf("%s%s", httpPrefixSchema, route.Spec.Host), nil
			}
			return fmt.Sprintf("%s%s", httpsPrefixSchema, route.Spec.Host), nil
		}
	}
	return "", nil
}
