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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/reference"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RaiseInfoEventf(obj runtime.Object, scheme *runtime.Scheme, c client.Client, reason, messageFormat string, args ...interface{}) error {
	return raiseEvent(obj, scheme, c, reason, fmt.Sprintf(messageFormat, args...), corev1.EventTypeNormal)
}

func RaiseWarnEventf(obj runtime.Object, scheme *runtime.Scheme, c client.Client, reason, messageFormat string, args ...interface{}) error {
	return raiseEvent(obj, scheme, c, reason, fmt.Sprintf(messageFormat, args...), corev1.EventTypeWarning)
}

func raiseEvent(obj runtime.Object, scheme *runtime.Scheme, c client.Client, reason, message, eventType string) error {
	ref, err := reference.GetReference(scheme, obj)
	if err != nil {
		return fmt.Errorf("unable to generate reference from object: %v", err)
	}
	event := newEvent(ref, reason, message, eventType)
	if err := c.Create(ctx.Background(), event); err != nil {
		return fmt.Errorf("unable to create event: %v", err)
	}
	return nil
}

func newEvent(ref *corev1.ObjectReference, reason, message, eventType string) *corev1.Event {
	t := metav1.Now()
	return &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%x", ref.Name, t.UnixNano()),
			Namespace: ref.Namespace,
		},
		InvolvedObject: *ref,
		Reason:         reason,
		Message:        message,
		Source:         corev1.EventSource{Component: ref.Name},
		FirstTimestamp: t,
		LastTimestamp:  t,
		Count:          1,
		Type:           eventType,
	}
}
