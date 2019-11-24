//     Copyright 2019 Nexus Operator and/or its authors
//
//     This file is part of Nexus Operator.
//
//     Nexus Operator is free software: you can redistribute it and/or modify
//     it under the terms of the GNU General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.
//
//     Nexus Operator is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU General Public License for more details.
//
//     You should have received a copy of the GNU General Public License
//     along with Nexus Operator.  If not, see <https://www.gnu.org/licenses/>.

package kubernetes

import (
	"context"
	"fmt"
	"github.com/m88i/nexus-operator/pkg/util"
	"k8s.io/api/networking/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetIngressURI discover the URI for Ingress
func GetIngressURI(cli client.Client, ingressName types.NamespacedName) (string, error) {
	ingress := &v1beta1.Ingress{}
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
