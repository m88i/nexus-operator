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

package meta

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/api/v1alpha1"
)

const AppLabel = "app"

func DefaultObjectMeta(nexus *v1alpha1.Nexus) v1.ObjectMeta {
	return v1.ObjectMeta{
		Namespace: nexus.Namespace,
		Name:      nexus.Name,
		Labels:    GenerateLabels(nexus),
	}
}

func GenerateLabels(nexus *v1alpha1.Nexus) map[string]string {
	nexusAppLabels := map[string]string{}
	nexusAppLabels[AppLabel] = nexus.Name
	return nexusAppLabels
}

func DefaultNetworkingMeta(nexus *v1alpha1.Nexus) v1.ObjectMeta {
	meta := DefaultObjectMeta(nexus)
	meta.Annotations = nexus.Spec.Networking.Annotations
	return meta
}
