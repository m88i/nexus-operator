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
	"github.com/m88i/aicura/nexus"
)

var communityMavenProxies = map[string]nexus.MavenProxyRepository{
	"apache":  defaultMavenProxyInstance("apache", "https://repo.maven.apache.org/maven2/"),
	"red-hat": defaultMavenProxyInstance("red-hat", "https://maven.repository.redhat.com/ga/"),
	"jboss":   defaultMavenProxyInstance("jboss", "https://repository.jboss.org/"),
}

const (
	mavenCentralRepoID = "maven-public"
)

// RepositoryOperations describes the public operations in the repository domain for the Nexus instance
type RepositoryOperations interface {
	EnsureCommunityMavenProxies() error
}

type repositoryOperation struct {
	server
}

func repositoryOperations(server *server) RepositoryOperations {
	return &repositoryOperation{server: *server}
}

func (r *repositoryOperation) EnsureCommunityMavenProxies() error {
	if r.nexus.Spec.ServerOperations.DisableRepositoryCreation {
		log.Debug("'spec.serverOperations.disableRepositoryCreation' is set to 'true'. Skipping repository creation")
		return nil
	}
	if err := r.createCommunityReposIfNotExists(); err != nil {
		return err
	}
	return r.addCommunityReposToMavenCentralGroup()
}

func (r *repositoryOperation) addCommunityReposToMavenCentralGroup() error {
	log.Debug("Attempt to fetch the Maven Central group repository")
	mavenCentral, err := r.nexuscli.MavenGroupRepositoryService.GetRepoByName(mavenCentralRepoID)
	if err != nil {
		return err
	}
	if mavenCentral == nil {
		log.Warnf("Maven Central repository group not found in the server instance, won't add community repos to the group")
		return nil
	}

	var newMembers []string
	for newMember := range communityMavenProxies {
		found := false
		for _, added := range mavenCentral.Group.MemberNames {
			if newMember == added {
				found = true
				break
			}
		}
		if !found {
			newMembers = append(newMembers, newMember)
		}
	}

	if len(newMembers) > 0 {
		log.Debugf("Community repositories to be added in the Maven Central group: %v", newMembers)
		mavenCentral.Group.MemberNames = append(mavenCentral.Group.MemberNames, newMembers...)

		err = r.nexuscli.MavenGroupRepositoryService.Update(*mavenCentral)
		if err == nil {
			log.Debug("Maven Central updated with new community members")
			r.status.MavenCentralUpdated = true
		}
		return err
	}
	log.Debug("Community repositories already added to the Maven Central repo")
	r.status.MavenCentralUpdated = true
	return nil
}

func (r *repositoryOperation) createCommunityReposIfNotExists() error {
	var reposToAdd []nexus.MavenProxyRepository
	log.Debug("Attempt to create community repositories")
	for _, repo := range communityMavenProxies {
		log.Debugf("Trying to fetch repository %s", repo.Name)
		fetchedRepo, err := r.nexuscli.MavenProxyRepositoryService.GetRepoByName(repo.Name)
		if err != nil {
			return err
		}
		if fetchedRepo == nil {
			reposToAdd = append(reposToAdd, repo)
		}
	}
	if len(reposToAdd) > 0 {
		log.Debugf("Repositories to add %v", reposToAdd)
		if err := r.nexuscli.MavenProxyRepositoryService.Add(reposToAdd...); err != nil {
			return err
		}
		log.Debug("All repositories created")
	}
	log.Debug("Community repositories already created, skipping")
	r.status.CommunityRepositoriesCreated = true
	return nil
}

func defaultMavenProxyInstance(name, url string) nexus.MavenProxyRepository {
	return nexus.MavenProxyRepository{
		Proxy: nexus.Proxy{
			MetadataMaxAge: 1440,
			RemoteURL:      url,
			ContentMaxAge:  -1,
		},
		Repository: nexus.Repository{
			Online: nexus.NewBool(true),
			Format: nexus.NewRepositoryFormat(nexus.RepositoryFormatMaven2),
			Name:   name,
			Type:   nexus.NewRepositoryType(nexus.RepositoryTypeProxy),
		},
		Storage: nexus.Storage{
			BlobStoreName:               "default",
			StrictContentTypeValidation: true,
		},
		NegativeCache: nexus.NegativeCache{
			Enabled:    true,
			TimeToLive: 1440,
		},
		Maven: nexus.Maven{
			VersionPolicy: nexus.VersionPolicyRelease,
			LayoutPolicy:  nexus.LayoutPolicyPermissive,
		},
		HTTPClient: nexus.HTTPClient{
			Blocked:   nexus.NewBool(false),
			AutoBlock: true,
		},
	}
}
