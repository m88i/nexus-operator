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
	"encoding/json"
	"testing"

	"github.com/m88i/aicura/nexus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func Test_repositoryOperation_setMavenPublicURL(t *testing.T) {
	expectedURL := "http://nexus3." + t.Name() + "/repository/maven-public/"
	server, _ := createNewServerAndKubeCli(t, &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "nexus3", Namespace: t.Name()}})
	operations := &repositoryOperation{server: *server}
	repository := &nexus.MavenGroupRepository{}
	repositoryJSON := "  {\n    \"name\": \"maven-public\",\n    \"format\": \"maven2\",\n    \"url\": \"http://localhost:8081/repository/maven-public\",\n    \"online\": true,\n    \"storage\": {\n      \"blobStoreName\": \"default\",\n      \"strictContentTypeValidation\": true\n    },\n    \"group\": {\n      \"memberNames\": [\n        \"maven-releases\",\n        \"maven-snapshots\",\n        \"maven-central\"\n      ]\n    },\n    \"type\": \"group\"\n  }"
	err := json.Unmarshal([]byte(repositoryJSON), repository)
	assert.NoError(t, err)
	err = operations.setMavenPublicURL(repository)
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, operations.status.MavenPublicURL)
}
