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

package persistence

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"reflect"
	"testing"
)

func TestNewManager_setDefaults(t *testing.T) {
	tests := []struct {
		name  string
		input *v1alpha1.Nexus
		want  *v1alpha1.Nexus
	}{
		{
			"'spec.persistence.volumeSize' left blank",
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Persistence: v1alpha1.NexusPersistence{Persistent: true}}},
			&v1alpha1.Nexus{Spec: v1alpha1.NexusSpec{Persistence: v1alpha1.NexusPersistence{Persistent: true, VolumeSize: defaultVolumeSize}}},
		},
	}

	for _, tt := range tests {
		manager := &manager{
			nexus: tt.input,
		}
		manager.setDefaults()
		if !reflect.DeepEqual(manager.nexus, tt.want) {
			t.Errorf("TestManager_setDefaults() - %s\nWant: %v\tGot: %v", tt.name, tt.want, *manager.nexus)
		}
	}
}
