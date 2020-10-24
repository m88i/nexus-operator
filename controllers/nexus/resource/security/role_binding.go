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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
)

const clusterRoleKind = "ClusterRole"

func defaultRoleBinding(nexus *v1alpha1.Nexus) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{Name: "community-nexus-uid-200"},
		Subjects: []rbacv1.Subject{
			{
				Name:      nexus.Spec.ServiceAccountName,
				Namespace: nexus.Namespace,
				Kind:      rbacv1.ServiceAccountKind,
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     clusterRoleKind,
			Name:     clusterRoleName,
		},
	}
}
