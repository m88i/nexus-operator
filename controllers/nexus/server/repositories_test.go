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
)

// TODO: add a test to verify the Maven Central group being updated with the new members. See: https://github.com/m88i/aicura/issues/18

func TestAddCommReposNoCentralGroup(t *testing.T) {
	server, _ := createNewServerAndKubeCli(t)
	err := repositoryOperations(server).EnsureCommunityMavenProxies()
	assert.NoError(t, err)
	repos, err := server.nexuscli.MavenProxyRepositoryService.List()
	assert.NoError(t, err)
	assert.Len(t, repos, len(communityMavenProxies))
}
