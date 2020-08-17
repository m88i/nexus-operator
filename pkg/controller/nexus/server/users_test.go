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

package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_userOperation_EnsureOperatorUser(t *testing.T) {
	server, _ := createNewServerAndKubeCli(t, &corev1.Secret{ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()}})

	err := userOperations(server).EnsureOperatorUser()
	assert.NoError(t, err)
	user, err := server.nexuscli.UserService.GetUserByID(operatorUsername)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, operatorUsername, user.UserID)
	assert.True(t, server.status.OperatorUserCreated)
}

func Test_userOperation_EnsureOperatorUser_AlreadyExists(t *testing.T) {
	server, _ := createNewServerAndKubeCli(t,
		&corev1.Secret{
			ObjectMeta: v1.ObjectMeta{Name: "nexus3", Namespace: t.Name()},
			Data: map[string][]byte{
				SecretKeyPassword: []byte("12345"),
				SecretKeyUsername: []byte(operatorUsername),
			}})

	err := userOperations(server).EnsureOperatorUser()
	assert.NoError(t, err)
	user, err := server.nexuscli.UserService.GetUserByID(operatorUsername)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, operatorUsername, user.UserID)
	assert.True(t, server.status.OperatorUserCreated)
}
