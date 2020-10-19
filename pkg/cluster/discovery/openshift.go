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
	routev1 "github.com/openshift/api/route/v1"
)

const (
	openshiftGroup = "openshift.io"
)

// IsOpenShift verify if the operator is running on OpenShift
func IsOpenShift() (bool, error) {
	return hasGroup(openshiftGroup)
}

// IsRouteAvailable verifies if the current cluster has the Route API from OpenShift available
func IsRouteAvailable() (bool, error) {
	return hasGroupVersion(routev1.GroupName, routev1.GroupVersion.Version)
}
