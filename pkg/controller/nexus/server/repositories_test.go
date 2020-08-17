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
