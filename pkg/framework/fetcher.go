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

package framework

import (
	ctx "context"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Fetch(client client.Client, key types.NamespacedName, instance resource.KubernetesResource) error {
	log.Debugf("Attempting to fetch deployed %s (%s)", instance.GetObjectKind(), key.Name)
	if err := client.Get(ctx.TODO(), key, instance); err != nil {
		if errors.IsNotFound(err) {
			log.Debugf("There is no deployed %s (%s)", instance.GetObjectKind(), key.Name)
		}
		return err
	}
	return nil
}
