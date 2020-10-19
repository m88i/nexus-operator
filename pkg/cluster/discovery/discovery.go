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

package discovery

import (
	"strings"

	"k8s.io/client-go/discovery"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

var cli discovery.DiscoveryInterface = discovery.NewDiscoveryClientForConfigOrDie(controllerruntime.GetConfigOrDie())

// SetClient sets the package-level Discovery client.
// You probably don't need this for production. Exists to facilitate testing.
func SetClient(disc discovery.DiscoveryInterface) {
	cli = disc
}

// hasGroup checks if the given group name is available in the cluster
func hasGroup(group string) (bool, error) {
	groups, err := cli.ServerGroups()
	if err != nil {
		return false, err
	}
	for _, g := range groups.Groups {
		if strings.Contains(g.Name, group) {
			return true, nil
		}
	}
	return false, nil
}

// hasGroupVersion checks if the given group name and version is available in the cluster
func hasGroupVersion(group, version string) (bool, error) {
	serverGroups, err := cli.ServerGroups()
	if err != nil {
		return false, err
	}
	for _, serverGroup := range serverGroups.Groups {
		if serverGroup.Name == group {
			for _, availableVersion := range serverGroup.Versions {
				if availableVersion.Version == version {
					return true, nil
				}
			}
			// we found the correct group, but not the correct version, so fail
			return false, nil
		}
	}
	return false, nil
}
