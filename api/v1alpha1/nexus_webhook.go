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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/m88i/nexus-operator/pkg/logger"
)

const admissionLogName = "admission"

// log is for logging in this package.
var log = logger.GetLogger(admissionLogName)

func (r *Nexus) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-apps-m88i-io-m88i-io-v1alpha1-nexus,mutating=true,failurePolicy=fail,groups=apps.m88i.io.m88i.io,resources=nexus,verbs=create;update,versions=v1alpha1,name=mnexus.kb.io

var _ webhook.Defaulter = &Nexus{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Nexus) Default() {
	log = logger.GetLoggerWithResource(admissionLogName, r)
	defer func() { log = logger.GetLogger(admissionLogName) }()

	log.Info("Setting defaults and making necessary changes to the Nexus")
	newMutator(r).mutate()
}

// if we ever need validation upon delete requests change verbs to "verbs=create;update;delete".
// +kubebuilder:webhook:verbs=create;update,path=/validate-apps-m88i-io-m88i-io-v1alpha1-nexus,mutating=false,failurePolicy=fail,groups=apps.m88i.io.m88i.io,resources=nexus,versions=v1alpha1,name=vnexus.kb.io

var _ webhook.Validator = &Nexus{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Nexus) ValidateCreate() error {
	log = logger.GetLoggerWithResource(admissionLogName, r)
	defer func() { log = logger.GetLogger(admissionLogName) }()

	log.Info("Validating a new Nexus")
	return newValidator(r).validate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Nexus) ValidateUpdate(old runtime.Object) error {
	log = logger.GetLoggerWithResource(admissionLogName, r)
	defer func() { log = logger.GetLogger(admissionLogName) }()

	log.Info("Validating an update to a existing Nexus")
	return newValidator(r).validate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Nexus) ValidateDelete() error {
	// we don't care about validation on delete requests
	return nil
}
