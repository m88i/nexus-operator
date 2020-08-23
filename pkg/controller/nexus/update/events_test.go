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

package update

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
