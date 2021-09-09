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

	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/pkg/util"
)

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
