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

package update

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/controller/nexus/event"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	successfulUpdateReason = "UpdateSuccess"
	failedUpdateReason     = "UpdateFailed"
)

func createUpdateSuccessEvent(nexus *v1alpha1.Nexus, scheme *runtime.Scheme, c client.Client, tag string) {
	err := event.RaiseInfoEventf(nexus, scheme, c, successfulUpdateReason, "Successfully updated to %s", tag)
	if err != nil {
		log.Warnf("Unable to raise event for successful update of %s to %s: %v", nexus.Name, tag, err)
	}
}

func createUpdateFailureEvent(nexus *v1alpha1.Nexus, scheme *runtime.Scheme, c client.Client, tag string) {
	err := event.RaiseWarnEventf(nexus, scheme, c, failedUpdateReason, "Failed to update to %s. Human intervention may be required", tag)
	if err != nil {
		log.Warnf("Unable to raise event for failed update of %s to %s: %v", nexus.Name, tag, err)
	}
}
