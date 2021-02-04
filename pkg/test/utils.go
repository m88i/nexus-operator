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
	"reflect"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/RHsyseng/operator-utils/pkg/resource"
)

// ContainsType returns true if the give resource slice contains an element of type t
func ContainsType(resources []resource.KubernetesResource, t reflect.Type) bool {
	for _, res := range resources {
		if reflect.TypeOf(res) == t {
			return true
		}
	}
	return false
}

// EventExists returns true if an event with the given reason exists
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

// IsInterfaceValueNil returns true if the value stored by the interface is nil
// See https://medium.com/@glucn/golang-an-interface-holding-a-nil-value-is-not-nil-bb151f472cc7
func IsInterfaceValueNil(i interface{}) bool {
	return i == nil || (reflect.ValueOf(i).Kind() == reflect.Ptr && reflect.ValueOf(i).IsNil())
}
