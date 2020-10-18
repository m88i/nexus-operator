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

package security

import (
	secv1 "github.com/openshift/api/security/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

var (
	communityUID = int64(200)
	sccName      = "allow-nexus-userid-200"
)

func defaultSCC() *secv1.SecurityContextConstraints {
	return &secv1.SecurityContextConstraints{
		ObjectMeta: controllerruntime.ObjectMeta{Name: sccName},
		FSGroup: secv1.FSGroupStrategyOptions{
			Type:   secv1.FSGroupStrategyMustRunAs,
			Ranges: []secv1.IDRange{{Min: communityUID, Max: communityUID}},
		},
		RunAsUser: secv1.RunAsUserStrategyOptions{
			Type: secv1.RunAsUserStrategyMustRunAs,
			UID:  &communityUID,
		},
		SELinuxContext: secv1.SELinuxContextStrategyOptions{
			Type: secv1.SELinuxStrategyMustRunAs,
		},
		SupplementalGroups: secv1.SupplementalGroupsStrategyOptions{
			Type:   secv1.SupplementalGroupsStrategyMustRunAs,
			Ranges: []secv1.IDRange{{Min: communityUID, Max: communityUID}},
		},
		Volumes: []secv1.FSType{
			secv1.FSTypePersistentVolumeClaim,
			secv1.FSTypeSecret,
		},
	}
}
