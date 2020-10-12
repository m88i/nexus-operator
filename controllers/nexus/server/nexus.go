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
	"context"
	"fmt"
	"os"

	nexusapi "github.com/m88i/aicura/nexus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
)

type server struct {
	nexus     *v1alpha1.Nexus
	k8sclient client.Client
	nexuscli  *nexusapi.Client
	status    *v1alpha1.OperationsStatus
}

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin123"
	// used when running the operator instance locally
	serverURLEnvKey = "NEXUS_SERVER_URL"
)

func handleServerOperations(nexus *v1alpha1.Nexus, client client.Client, nexusAPIBuilder func(url, user, pass string) *nexusapi.Client) (v1alpha1.OperationsStatus, error) {
	s := server{nexus: nexus, k8sclient: client, status: &v1alpha1.OperationsStatus{}}
	if nexus.Spec.GenerateRandomAdminPassword {
		return *s.status, nil
	}
	log.Debug("Initializing server operations", "instance", nexus.Name)
	if s.isServerReady() {
		internalEndpoint, err := s.getNexusEndpoint()
		if err != nil {
			s.status.Reason = fmt.Sprintf("Impossible to resolve endpoint for Nexus instance %s. Error: %s", nexus.Name, err.Error())
			s.status.ServerReady = false
			return *s.status, nil
		}
		s.nexuscli = nexusAPIBuilder(internalEndpoint, defaultAdminUsername, defaultAdminPassword)

		if err := userOperations(&s).EnsureOperatorUser(); err != nil {
			s.status.Reason = err.Error()
			return *s.status, err
		}
		if err := repositoryOperations(&s).EnsureCommunityMavenProxies(); err != nil {
			s.status.Reason = err.Error()
			return *s.status, err
		}
		s.status.Reason = ""
	}
	return *s.status, nil
}

// HandleServerOperations makes all required operations in the Nexus server side, such as creating the operator user
func HandleServerOperations(nexus *v1alpha1.Nexus, client client.Client) (v1alpha1.OperationsStatus, error) {
	log = logger.GetLoggerWithResource(defaultLogName, nexus)
	defer func() { log = logger.GetLogger(defaultLogName) }()
	return handleServerOperations(nexus, client, func(url, user, pass string) *nexusapi.Client {
		return nexusapi.NewClient(url).WithCredentials(user, pass).Build()
	})
}

func (s *server) getNexusEndpoint() (string, error) {
	externalURL := os.Getenv(serverURLEnvKey)
	if len(externalURL) > 0 {
		return externalURL, nil
	}

	svc := &corev1.Service{}
	if err := s.k8sclient.Get(context.TODO(), types.NamespacedName{Name: s.nexus.Name, Namespace: s.nexus.Namespace}, svc); err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s.%s", svc.Name, svc.Namespace), nil
}

// isServerReady checks if the given Nexus instance is ready to receive requests
func (s *server) isServerReady() bool {
	if s.nexus.Status.DeploymentStatus.AvailableReplicas > 0 {
		s.status.ServerReady = true
		s.status.Reason = ""
		return true
	}
	s.status.ServerReady = false
	s.status.Reason = "Server does not have enough available replicas"
	return false
}
