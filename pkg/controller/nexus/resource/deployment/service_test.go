//     Copyright 2019 Nexus Operator and/or its authors
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

package deployment

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/framework"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_newService(t *testing.T) {
	appName := "nexus3"
	nexus := &v1alpha1.Nexus{
		ObjectMeta: v1.ObjectMeta{
			Name:      appName,
			Namespace: t.Name(),
		},
		Spec: v1alpha1.NexusSpec{
			Replicas: 1,
			Persistence: v1alpha1.NexusPersistence{
				Persistent: false,
			},
		},
	}
	svc := newService(nexus)

	assert.Len(t, svc.Spec.Ports, 1)
	assert.Equal(t, int32(NexusServicePort), svc.Spec.Ports[0].Port)
	assert.Equal(t, appName, svc.Labels[framework.AppLabel])
	assert.Equal(t, appName, svc.Spec.Selector[framework.AppLabel])
}
