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
	"fmt"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/api/v1alpha1"
)

// FetchDeployedResources fetches deployed resources whose Kind is present in "managedObjectsRef"
func FetchDeployedResources(managedObjectsRef map[string]resource.KubernetesResource, nexus *v1alpha1.Nexus, cli client.Client) ([]resource.KubernetesResource, error) {
	var resources []resource.KubernetesResource
	for resKind, resRef := range managedObjectsRef {
		key := Key(nexus)
		if err := Fetch(cli, key, resRef, resKind); err == nil {
			resources = append(resources, resRef)
		} else if !errors.IsNotFound(err) {
			return nil, fmt.Errorf("could not fetch %s (%s/%s): %v", resKind, nexus.Namespace, nexus.Name, err)
		} else {
			log.Debug("Unable to find resource", "kind", resKind, "namespacedName", key)
		}
	}
	return resources, nil
}

// Fetch fetches a single deployed resource and stores it in "instance"
func Fetch(client client.Client, key types.NamespacedName, instance resource.KubernetesResource, kind string) error {
	log.Info("Attempting to fetch deployed resource", "kind", kind, "namespacedName", key)
	if err := client.Get(ctx.TODO(), key, instance); err != nil {
		if errors.IsNotFound(err) {
			log.Debug("Unable to find resource", "kind", kind, "namespacedName", key)
		}
		return err
	}
	return nil
}
