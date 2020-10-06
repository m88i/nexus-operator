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

package validation

import (
	"github.com/m88i/nexus-operator/pkg/apis/apps/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/cluster/kubernetes"
	"github.com/m88i/nexus-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const changedNexusReason = "NexusSpecChanged"

func createChangedNexusEvent(nexus *v1alpha1.Nexus, scheme *runtime.Scheme, c client.Client, field string) {
	log := logger.GetLoggerWithResource("validation_event", nexus)
	err := kubernetes.RaiseWarnEventf(nexus, scheme, c, changedNexusReason, "'%s' has been changed in %s/%s. Check the logs for more information", field, nexus.Namespace, nexus.Name)
	if err != nil {
		log.Warnf("Unable to raise event for changing '%s' in Nexus CR: %v", field, err)
	}
}
