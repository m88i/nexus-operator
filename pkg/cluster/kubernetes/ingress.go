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

package kubernetes

import (
	"context"
	"fmt"

	"github.com/m88i/nexus-operator/pkg/util"
	networking "k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const ingressGroup = networking.GroupName

// It would be nice to keep the version as constant as well, but the package only offers it as a variable.
// FIXME if this ever changes and turn this into a const
var ingressVersion = networking.SchemeGroupVersion.Version

// IsIngressAvailable checks if th cluster supports Ingresses from k8s.io/api/networking/v1beta1
// <lcaparelli> TODO: consider an implementation on which callers don't need a discovery interface and that caches results from calls (these don't usually change, so a high TTL should be ok). Same applies to similar functions on this package
func IsIngressAvailable(d discovery.DiscoveryInterface) (bool, error) {
	serverGroups, err := d.ServerGroups()
	if err != nil {
		return false, err
	}
	for _, serverGroup := range serverGroups.Groups {
		if serverGroup.Name == ingressGroup {
			for _, version := range serverGroup.Versions {
				if version.Version == ingressVersion {
					return true, nil
				}
			}
			// we found the correct group, but not the correct version, so fail
			return false, nil
		}
	}
	return false, nil
}

// GetIngressURI discover the URI for Ingress
func GetIngressURI(cli client.Client, ingressName types.NamespacedName) (string, error) {
	ingress := &networking.Ingress{}
	if err := cli.Get(context.TODO(), ingressName, ingress); err != nil && !errors.IsNotFound(err) {
		return "", err
	} else if errors.IsNotFound(err) {
		return "", nil
	}

	if len(ingress.Spec.Rules) > 0 {
		if len(ingress.Spec.TLS) == 0 {
			return fmt.Sprintf("%s%s", util.HTTPPrefixSchema, ingress.Spec.Rules[0].Host), nil
		}
		return fmt.Sprintf("%s%s", util.HTTPSPrefixSchema, ingress.Spec.Rules[0].Host), nil
	}

	return "", nil
}
