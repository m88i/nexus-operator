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
