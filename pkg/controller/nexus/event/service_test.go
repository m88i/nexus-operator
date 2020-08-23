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

package event

import (
	ctx "context"
	"fmt"
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
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
