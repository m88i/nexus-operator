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

package test

import (
	"context"
	"k8s.io/api/core/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/RHsyseng/operator-utils/pkg/resource"
)

func ContainsType(resources []resource.KubernetesResource, t reflect.Type) bool {
	for _, res := range resources {
		if reflect.TypeOf(res) == t {
			return true
		}
	}
	return false
}

func EventExists(c client.Client, reason string) bool {
	eventList := &v1.EventList{}
	_ = c.List(context.Background(), eventList)
	for _, event := range eventList.Items {
		if event.Reason == reason {
			return true
		}
	}
	return false
}
