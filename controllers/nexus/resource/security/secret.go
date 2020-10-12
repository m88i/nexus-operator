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
	corev1 "k8s.io/api/core/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/controllers/nexus/resource/meta"
)

// defaultSecret all purposes secret used by this instance
func defaultSecret(nexus *v1alpha1.Nexus) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: meta.DefaultObjectMeta(nexus),
		Type:       corev1.SecretTypeOpaque,
	}
}
