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
	rbacv1 "k8s.io/api/rbac/v1"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

const (
	verbUse         = "use"
	sccResourceName = "securitycontextconstraints"
	clusterRoleName = "nexus-community"
)

func defaultClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: controllerruntime.ObjectMeta{Name: clusterRoleName},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups:     []string{secv1.GroupName},
				ResourceNames: []string{sccName},
				Resources:     []string{sccResourceName},
				Verbs:         []string{verbUse},
			},
		},
	}
}
