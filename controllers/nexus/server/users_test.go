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
