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

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/m88i/nexus-operator/pkg/test"
	"github.com/stretchr/testify/assert"
)

func TestFetch(t *testing.T) {
	deployment := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: "deployment", Namespace: t.Name()}}
	cli := test.NewFakeClientBuilder(deployment).Build()
	err := Fetch(cli, Key(deployment), deployment)
	assert.NoError(t, err)
}

func TestNotFoundFetch(t *testing.T) {
	deployment := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: "deployment", Namespace: t.Name()}}
	cli := test.NewFakeClientBuilder().Build()
	err := Fetch(cli, Key(deployment), deployment)
	assert.Error(t, err)
	assert.True(t, errors.IsNotFound(err))
}
