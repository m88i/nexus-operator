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

// The update package manages logic related to automatic updates.
package update

import (
	ctx "context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
)

const (
	updateStartPrefix  = "Starting automatic update from"
	updateOKPrefix     = "Successfully updated from"
	updateFailedPrefix = "Failed to update from"

	updateStartFormat  = updateStartPrefix + " %s to %s"
	updateOKFormat     = updateOKPrefix + " %s to %s"
	updateFailedFormat = updateFailedPrefix + " %s to %s"
)

// HandleUpdate constructs state from 'nexus.status.updateConditions' and, based on this state, it may:
//   - mark an update as started
//   - mark an update as finished
// If an update fails automatic updates are disabled and the image is set to the previously deployed tag
//
// This is a state machine with two states: "idle" and "updating".
// "idle" transitions into "updating" if isNewUpdate == true.
// "updating" transitions back to "idle" if automatic updates get disabled or if the update fails/succeeds.
// "updating" transitions to itself if isNewUpdate == true.
func HandleUpdate(nexus *v1alpha1.Nexus, deployed, required *appsv1.Deployment, scheme *runtime.Scheme, c client.Client) error {
	if nexus.Spec.AutomaticUpdate.Disabled || differentImagesOrMinors(deployed, required) {
		if alreadyUpdating(nexus) {
			// we were in an update, so let's clear its status
			nexus.Status.UpdateConditions = nil
		}
		return nil
	}

	log = logger.GetLoggerWithResource(monitorLogName, nexus)
	defer func() { log = logger.GetLogger(defaultLogName) }()

	// it's important to check if this is a new update before checking ongoing updates because
	// if this is a new update, the one that was happening before no longer matters
	// so we just reset the update state and return
	if newUpdate, previousTag, targetTag := isNewUpdate(deployed, required); newUpdate {
		log.Info("Started tags update", "previous", previousTag, "target", targetTag)
		// the Nexus status update can be delayed, let's leave it to the reconciler
		nexus.Status.UpdateConditions = []string{fmt.Sprintf(updateStartFormat, previousTag, targetTag)}
		return nil
	}

	if !alreadyUpdating(nexus) {
		log.Debug("No ongoing update, nothing to check")
		// nothing to monitor, let's return
		return nil
	}

	previousTag, targetTag, err := getUpdateTags(nexus)
	if err != nil {
		log.Error(err, "Failed to parse 'status.updateConditions'. Was it tampered with? Unable to monitor ongoing update, human intervention may be required.", "deployment", deployed.Name)
		nexus.Status.UpdateConditions = nil
		return nil
	}

	for _, condition := range deployed.Status.Conditions {
		if condition.Type == appsv1.DeploymentProgressing && condition.Status == "False" {
			log.Warn("Update failed: Human intervention may be required", "target tag", targetTag, "Reason", condition.Reason, "Message", condition.Message)
			nexus.Status.UpdateConditions = append(nexus.Status.UpdateConditions, fmt.Sprintf(updateFailedFormat, previousTag, targetTag))

			// we must return an error if we can't disable automatic updates
			// this can't be delayed like the status updates as need the reconcile request to be requeued
			if err := rollback(nexus, previousTag, c); err != nil {
				return fmt.Errorf("the update has failed, but could not disable automatic updates: %v", err)
			}

			// we don't want to create spurious events, so we only raise it after we've disabled updates
			// and we know this part of the function won't be reached again
			createUpdateFailureEvent(nexus, scheme, c, targetTag)
			return nil
		}

		if condition.Type == appsv1.DeploymentProgressing && condition.Reason == "NewReplicaSetAvailable" {
			log.Info("Successfully updated", "tag", targetTag)
			// the Nexus status update can be delayed, let's leave it to the reconciler
			nexus.Status.UpdateConditions = append(nexus.Status.UpdateConditions, fmt.Sprintf(updateOKFormat, previousTag, targetTag))
			createUpdateSuccessEvent(nexus, scheme, c, targetTag)
			return nil
		}
	}
	return nil
}

func alreadyUpdating(nexus *v1alpha1.Nexus) bool {
	if len(nexus.Status.UpdateConditions) == 0 {
		return false
	}
	lastCondition := nexus.Status.UpdateConditions[len(nexus.Status.UpdateConditions)-1]
	if strings.HasPrefix(lastCondition, updateFailedPrefix) || strings.HasPrefix(lastCondition, updateOKPrefix) {
		return false
	}
	return true
}

func isNewUpdate(deployed, required *appsv1.Deployment) (updating bool, previousTag, targetTag string) {
	depImage := deployed.Spec.Template.Spec.Containers[0].Image
	reqImage := required.Spec.Template.Spec.Containers[0].Image
	deployedImageParts := strings.Split(depImage, ":")
	requiredImageParts := strings.Split(reqImage, ":")

	updating, err := HigherVersion(requiredImageParts[1], deployedImageParts[1])
	if err != nil {
		log.Error(err, "Unable to check if the required Deployment is an update when comparing to the deployed one", "deployment", required.Name)
		return
	}
	previousTag = deployedImageParts[1]
	targetTag = requiredImageParts[1]
	return
}

func getUpdateTags(nexus *v1alpha1.Nexus) (previousTag, targetTag string, err error) {
	_, err = fmt.Sscanf(nexus.Status.UpdateConditions[0], updateStartFormat, &previousTag, &targetTag)
	return
}

func differentImagesOrMinors(deployed, required *appsv1.Deployment) bool {
	depImage := deployed.Spec.Template.Spec.Containers[0].Image
	reqImage := required.Spec.Template.Spec.Containers[0].Image
	requiredImageParts := strings.Split(reqImage, ":")
	deployedImageParts := strings.Split(depImage, ":")

	// different images, not an update
	if requiredImageParts[0] != deployedImageParts[0] {
		return true
	}

	// Might be the same, but we can't tell, so let's be conservative and say it isn't
	if len(deployedImageParts) == 1 || deployedImageParts[1] == "latest" {
		return true
	}

	// we should be able to assume there will be no parsing error, we just created this deployment in the reconcile loop
	reqMinor, _ := getMinor(requiredImageParts[1])
	// the deployed one, on the other hand, might have been tampered with
	depMinor, err := getMinor(deployedImageParts[1])
	if err != nil {
		log.Error(err, "Unable to parse the deployed Deployment's tag. Cannot determine if this is an update. Has it been tampered with?", "deployment", deployed.Name)
		return true
	}

	return reqMinor != depMinor
}

func rollback(nexus *v1alpha1.Nexus, tag string, c client.Client) error {
	// disable automatic updates
	nexus.Spec.AutomaticUpdate.Disabled = true
	nexus.Spec.AutomaticUpdate.MinorVersion = nil
	// Let's set the tag to one we know is working.
	nexus.Spec.Image = fmt.Sprintf("%s:%s", strings.Split(nexus.Spec.Image, ":")[0], tag)
	return c.Update(ctx.Background(), nexus)
}
