//     Copyright 2020 Nexus Operator and/or its authors
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

package framework

import (
	ctx "context"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Fetch(client client.Client, key types.NamespacedName, instance resource.KubernetesResource) error {
	log.Debugf("Attempting to fetch deployed %s (%s)", instance.GetObjectKind(), instance.GetName())
	if err := client.Get(ctx.TODO(), key, instance); err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed %s (%s)", instance.GetObjectKind(), instance.GetName())
		}
		return err
	}
	return nil
}
