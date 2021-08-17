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

package controllers

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/m88i/nexus-operator/api/v1alpha1"
	"github.com/m88i/nexus-operator/pkg/logger"
)

var _ = Describe("Nexus Controller", func() {
	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		NexusName      = "nexus3"
		NexusNamespace = "default"

		cpu         = "2"
		memory      = "2Gi"
		exposedPort = 31031

		timeout  = time.Minute * 1
		interval = time.Second * 10
	)

	var (
		log          = logger.GetLogger("integration-test")
		defaultNexus = &v1alpha1.Nexus{
			ObjectMeta: metav1.ObjectMeta{
				Name:      NexusName,
				Namespace: NexusNamespace,
			},
			Spec: v1alpha1.NexusSpec{
				Replicas:       1,
				UseRedHatImage: false,
				Resources: v1.ResourceRequirements{
					Limits: map[v1.ResourceName]resource.Quantity{
						v1.ResourceCPU:    resource.MustParse(cpu),
						v1.ResourceMemory: resource.MustParse(memory),
					},
				},
				Persistence: v1alpha1.NexusPersistence{Persistent: false},
				Networking: v1alpha1.NexusNetworking{
					Expose:   true,
					ExposeAs: v1alpha1.NodePortExposeType,
					NodePort: exposedPort,
				},
				Properties: map[string]string{
					"nexus.properties.1": "value1",
					"nexus.properties.2": "value2",
				},
			},
		}
	)

	Context("When creating a simple Nexus3 instance with no persistence, community image and NodePort exposed", func() {
		It("Should have nexusStatus set to Pending", func() {
			ctx := context.Background()
			nexus := defaultNexus.DeepCopy()

			Expect(k8sClient.Create(ctx, nexus)).Should(Succeed())

			deployment := &appsv1.Deployment{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: NexusName, Namespace: NexusNamespace}, deployment)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			configMap := &v1.ConfigMap{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: NexusName, Namespace: NexusNamespace}, configMap)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// there's no controllers deployed in this environment, so the Deployment would never have a pod deployed,
			// status would never change to "OK".
			// We basically verify if the controller is reconciling correct, status will be always Pending
			createdNexus := &v1alpha1.Nexus{}
			Eventually(func() bool {
				log.Info("Waiting for nexus instance to have ", "status", v1alpha1.NexusStatusPending)
				err := k8sClient.Get(ctx, types.NamespacedName{Name: NexusName, Namespace: NexusNamespace}, createdNexus)
				if errors.IsNotFound(err) {
					return false
				} else if err != nil {
					log.Error(err, "Failed to fetch Nexus instance", "Name", NexusName, "Namespace", NexusNamespace)
					return false
				}
				log.Info("Returning after fetching Nexus instance", "status", createdNexus.Status.NexusStatus)
				return createdNexus.Status.NexusStatus == v1alpha1.NexusStatusPending
			}, timeout, interval).Should(BeTrue())
		})
	})

})
