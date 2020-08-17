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

package framework

import (
	"testing"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAlwaysTrueComparator(t *testing.T) {
	deployment1 := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: "deployment1", Namespace: t.Name()}}
	deployment2 := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: "deployment2", Namespace: t.Name()}}
	assert.True(t, AlwaysTrueComparator()(deployment1, deployment2))
}
