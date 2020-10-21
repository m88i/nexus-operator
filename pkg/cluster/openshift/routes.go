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

package openshift

import (
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/pkg/cluster/discovery"
	"github.com/m88i/nexus-operator/pkg/util"
)

// GetRouteURI will discover the route scheme based on the given namespaced name for the route
func GetRouteURI(client client.Client, routeName types.NamespacedName) (uri string, err error) {
	routeAvailable, err := discovery.IsRouteAvailable()
	if err != nil || !routeAvailable {
		return "", err
	}

	route := &routev1.Route{}
	if err := client.Get(context.TODO(), routeName, route); err != nil {
		if !errors.IsNotFound(err) {
			return "", err
		}
	}

	if len(route.Spec.Host) > 0 {
		if route.Spec.TLS != nil {
			return fmt.Sprintf("%s%s", util.HTTPSPrefixSchema, route.Spec.Host), nil
		}
		return fmt.Sprintf("%s%s", util.HTTPPrefixSchema, route.Spec.Host), nil
	}

	return "", nil
}
