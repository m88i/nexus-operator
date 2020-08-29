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
	ctx "context"
	"fmt"
	"testing"

	"github.com/m88i/nexus-operator/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestRaiseInfoEventf(t *testing.T) {
	nexus := &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: "test"}}
	client := test.NewFakeClientBuilder().Build()
	reason := "test-reason"
	format := "%s %s"
	message := "test-message"
	extraArg := "extra"

	assert.NoError(t, RaiseInfoEventf(nexus, client.Scheme(), client, reason, format, message, extraArg))
	eventList := &corev1.EventList{}
	assert.NoError(t, client.List(ctx.TODO(), eventList))
	event := eventList.Items[0]
	assert.Equal(t, nexus.Name, event.Source.Component)
	assert.Equal(t, reason, event.Reason)
	assert.Equal(t, fmt.Sprintf(format, message, extraArg), event.Message)
	assert.Equal(t, corev1.EventTypeNormal, event.Type)
}

func TestRaiseWarnEventf(t *testing.T) {
	nexus := &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: "test"}}
	client := test.NewFakeClientBuilder().Build()
	reason := "test-reason"
	format := "%s %s"
	message := "test-message"
	extraArg := "extra"

	assert.NoError(t, RaiseWarnEventf(nexus, client.Scheme(), client, reason, format, message, extraArg))
	eventList := &corev1.EventList{}
	assert.NoError(t, client.List(ctx.TODO(), eventList))
	event := eventList.Items[0]
	assert.Equal(t, nexus.Name, event.Source.Component)
	assert.Equal(t, reason, event.Reason)
	assert.Equal(t, fmt.Sprintf(format, message, extraArg), event.Message)
	assert.Equal(t, corev1.EventTypeWarning, event.Type)
}

func TestServerFailure(t *testing.T) {
	nexus := &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: "test"}}
	client := test.NewFakeClientBuilder().Build()
	reason := "test-reason"
	message := "test-message"

	client.SetMockErrorForOneRequest(fmt.Errorf("mock-error"))
	assert.Error(t, RaiseInfoEventf(nexus, client.Scheme(), client, reason, message))
}

func TestReferenceFailure(t *testing.T) {
	nexus := &v1alpha1.Nexus{ObjectMeta: metav1.ObjectMeta{Name: "nexus", Namespace: "test"}}
	client := test.NewFakeClientBuilder().Build()
	reason := "test-reason"
	message := "test-message"

	// let's pass in the default scheme
	assert.Error(t, RaiseInfoEventf(nexus, runtime.NewScheme(), client, reason, message))
}
