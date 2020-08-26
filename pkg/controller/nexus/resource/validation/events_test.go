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

package validation

import (
	ctx "context"
	"fmt"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_createChangedNexusEvent(t *testing.T) {
	nexus := &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: "test"}}
	client := test.NewFakeClientBuilder().Build()
	field := "some-field"

	// first, let's test a failure
	client.SetMockErrorForOneRequest(fmt.Errorf("mock err"))
	createChangedNexusEvent(nexus, client.Scheme(), client, field)
	eventList := &corev1.EventList{}
	_ = client.List(ctx.TODO(), eventList)
	assert.Len(t, eventList.Items, 0)

	// now a successful one
	createChangedNexusEvent(nexus, client.Scheme(), client, field)
	_ = client.List(ctx.TODO(), eventList)
	assert.Len(t, eventList.Items, 1)
	event := eventList.Items[0]
	assert.Equal(t, changedNexusReason, event.Reason)
}
