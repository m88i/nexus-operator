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
