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

package update

import (
	ctx "context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
)

func TestCreateUpdateSuccessEvent(t *testing.T) {
	nexus := &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: "test"}}
	client := test.NewFakeClientBuilder().Build()
	tag := "3.25.0"

	// first, let's test a failure
	client.SetMockErrorForOneRequest(fmt.Errorf("mock err"))
	createUpdateSuccessEvent(nexus, client.Scheme(), client, tag)
	eventList := &corev1.EventList{}
	_ = client.List(ctx.TODO(), eventList)
	assert.Len(t, eventList.Items, 0)

	// now a successful one
	createUpdateSuccessEvent(nexus, client.Scheme(), client, tag)
	_ = client.List(ctx.TODO(), eventList)
	assert.Len(t, eventList.Items, 1)
	event := eventList.Items[0]
	assert.Equal(t, successfulUpdateReason, event.Reason)
}

func TestCreateUpdateFailureEvent(t *testing.T) {
	nexus := &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: "test"}}
	client := test.NewFakeClientBuilder().Build()
	tag := "3.25.0"

	// first, let's test a failure
	client.SetMockErrorForOneRequest(fmt.Errorf("mock err"))
	createUpdateFailureEvent(nexus, client.Scheme(), client, tag)
	eventList := &corev1.EventList{}
	_ = client.List(ctx.TODO(), eventList)
	assert.Len(t, eventList.Items, 0)

	// now a successful one
	createUpdateFailureEvent(nexus, client.Scheme(), client, tag)
	_ = client.List(ctx.TODO(), eventList)
	assert.Len(t, eventList.Items, 1)
	event := eventList.Items[0]
	assert.Equal(t, failedUpdateReason, event.Reason)
}
