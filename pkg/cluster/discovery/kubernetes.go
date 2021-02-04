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
	networkingv1 "k8s.io/api/networking/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"

	"github.com/m88i/nexus-operator/pkg/framework/kind"
)

// IsIngressAvailable checks if the cluster supports Ingresses from k8s.io/api/networking/v1
func IsIngressAvailable() (bool, error) {
	return hasGroupVersionKind(networkingv1.SchemeGroupVersion.Group, networkingv1.SchemeGroupVersion.Version, kind.IngressKind)
}

// IsLegacyIngressAvailable checks if the cluster supports Ingresses from k8s.io/api/networking/v1beta1
func IsLegacyIngressAvailable() (bool, error) {
	return hasGroupVersionKind(networkingv1beta1.SchemeGroupVersion.Group, networkingv1beta1.SchemeGroupVersion.Version, kind.IngressKind)
}

// IsAnyIngressAvailable checks if either v1 or v1beta1 ingresses are available
func IsAnyIngressAvailable() (bool, error) {
	legacyIngressAvailable, errLegacy := IsLegacyIngressAvailable()
	ingressAvailable, err := IsIngressAvailable()

	// both calls failed, let's error out
	if errLegacy != nil && err != nil {
		return false, errLegacy
	}

	// at least one call didn't fail
	return legacyIngressAvailable || ingressAvailable, nil
}
