// +build !ignore_autogenerated

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
// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/m88i/nexus-operator/api/v1alpha1.NexusPersistence": schema_m88i_nexus_operator_api_v1alpha1_NexusPersistence(ref),
		"github.com/m88i/nexus-operator/api/v1alpha1.NexusProbe":       schema_m88i_nexus_operator_api_v1alpha1_NexusProbe(ref),
		"github.com/m88i/nexus-operator/api/v1alpha1.NexusSpec":        schema_m88i_nexus_operator_api_v1alpha1_NexusSpec(ref),
		"github.com/m88i/nexus-operator/api/v1alpha1.NexusStatus":      schema_m88i_nexus_operator_api_v1alpha1_NexusStatus(ref),
	}
}

func schema_m88i_nexus_operator_api_v1alpha1_NexusPersistence(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "NexusPersistence is the structure for the data persistent",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"persistent": {
						SchemaProps: spec.SchemaProps{
							Description: "Flag to indicate if this instance will be persistent or not",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"volumeSize": {
						SchemaProps: spec.SchemaProps{
							Description: "If persistent, the size of the Volume. Defaults: 10Gi",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"storageClass": {
						SchemaProps: spec.SchemaProps{
							Description: "StorageClass used by the managed PVC.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
				},
				Required: []string{"persistent"},
			},
		},
	}
}

func schema_m88i_nexus_operator_api_v1alpha1_NexusProbe(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "NexusProbe describes a health check to be performed against a container to determine whether it is alive or ready to receive traffic.",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"initialDelaySeconds": {
						SchemaProps: spec.SchemaProps{
							Description: "Number of seconds after the container has started before probes are initiated. Defaults to 240 seconds. Minimum value is 0.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"timeoutSeconds": {
						SchemaProps: spec.SchemaProps{
							Description: "Number of seconds after which the probe times out. Defaults to 15 seconds. Minimum value is 1.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"periodSeconds": {
						SchemaProps: spec.SchemaProps{
							Description: "How often (in seconds) to perform the probe. Defaults to 10 seconds. Minimum value is 1.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"successThreshold": {
						SchemaProps: spec.SchemaProps{
							Description: "Minimum consecutive successes for the probe to be considered successful after having failed. Defaults to 1. Must be 1 for liveness and startup. Minimum value is 1.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"failureThreshold": {
						SchemaProps: spec.SchemaProps{
							Description: "Minimum consecutive failures for the probe to be considered failed after having succeeded. Defaults to 3. Minimum value is 1.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
				},
			},
		},
	}
}

func schema_m88i_nexus_operator_api_v1alpha1_NexusSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "NexusSpec defines the desired state of Nexus",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"replicas": {
						SchemaProps: spec.SchemaProps{
							Description: "Number of pod replicas desired. Defaults to 0.",
							Type:        []string{"integer"},
							Format:      "int32",
						},
					},
					"image": {
						SchemaProps: spec.SchemaProps{
							Description: "Full image tag name for this specific deployment. Will be ignored if `spec.useRedHatImage` is set to `true`. Default: docker.io/sonatype/nexus3:latest",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"automaticUpdate": {
						SchemaProps: spec.SchemaProps{
							Description: "Automatic updates configuration",
							Ref:         ref("github.com/m88i/nexus-operator/api/v1alpha1.NexusAutomaticUpdate"),
						},
					},
					"imagePullPolicy": {
						SchemaProps: spec.SchemaProps{
							Description: "The image pull policy for the Nexus image. If left blank behavior will be determined by the image tag (`Always` if \"latest\" and `IfNotPresent` otherwise). Possible values: `Always`, `IfNotPresent` or `Never`.",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"resources": {
						SchemaProps: spec.SchemaProps{
							Description: "Defined Resources for the Nexus instance",
							Ref:         ref("k8s.io/api/core/v1.ResourceRequirements"),
						},
					},
					"persistence": {
						SchemaProps: spec.SchemaProps{
							Description: "Persistence definition",
							Ref:         ref("github.com/m88i/nexus-operator/api/v1alpha1.NexusPersistence"),
						},
					},
					"useRedHatImage": {
						SchemaProps: spec.SchemaProps{
							Description: "If you have access to Red Hat Container Catalog, set this to `true` to use the certified image provided by Sonatype Defaults to `false`",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"generateRandomAdminPassword": {
						SchemaProps: spec.SchemaProps{
							Description: "GenerateRandomAdminPassword enables the random password generation. Defaults to `false`: the default password for a newly created instance is 'admin123', which should be changed in the first login. If set to `true`, you must use the automatically generated 'admin' password, stored in the container's file system at `/nexus-data/admin.password`. The operator uses the default credentials to create a user for itself to create default repositories. If set to `true`, the repositories won't be created since the operator won't fetch for the random password.",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"networking": {
						SchemaProps: spec.SchemaProps{
							Description: "Networking definition",
							Ref:         ref("github.com/m88i/nexus-operator/api/v1alpha1.NexusNetworking"),
						},
					},
					"serviceAccountName": {
						SchemaProps: spec.SchemaProps{
							Description: "ServiceAccountName is the name of the ServiceAccount used to run the Pods. If left blank, a default ServiceAccount is created with the same name as the Nexus CR (`metadata.name`).",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"livenessProbe": {
						SchemaProps: spec.SchemaProps{
							Description: "LivenessProbe describes how the Nexus container liveness probe should work",
							Ref:         ref("github.com/m88i/nexus-operator/api/v1alpha1.NexusProbe"),
						},
					},
					"readinessProbe": {
						SchemaProps: spec.SchemaProps{
							Description: "ReadinessProbe describes how the Nexus container readiness probe should work",
							Ref:         ref("github.com/m88i/nexus-operator/api/v1alpha1.NexusProbe"),
						},
					},
					"serverOperations": {
						SchemaProps: spec.SchemaProps{
							Description: "ServerOperations describes the options for the operations performed on the deployed server instance",
							Ref:         ref("github.com/m88i/nexus-operator/api/v1alpha1.ServerOperationsOpts"),
						},
					},
				},
				Required: []string{"replicas", "persistence", "useRedHatImage"},
			},
		},
		Dependencies: []string{
			"github.com/m88i/nexus-operator/api/v1alpha1.NexusAutomaticUpdate", "github.com/m88i/nexus-operator/api/v1alpha1.NexusNetworking", "github.com/m88i/nexus-operator/api/v1alpha1.NexusPersistence", "github.com/m88i/nexus-operator/api/v1alpha1.NexusProbe", "github.com/m88i/nexus-operator/api/v1alpha1.ServerOperationsOpts", "k8s.io/api/core/v1.ResourceRequirements"},
	}
}

func schema_m88i_nexus_operator_api_v1alpha1_NexusStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "NexusStatus defines the observed state of Nexus",
				Type:        []string{"object"},
				Properties: map[string]spec.Schema{
					"deploymentStatus": {
						SchemaProps: spec.SchemaProps{
							Description: "Condition status for the Nexus deployment",
							Ref:         ref("k8s.io/api/apps/v1.DeploymentStatus"),
						},
					},
					"nexusStatus": {
						SchemaProps: spec.SchemaProps{
							Description: "Will be \"OK\" when this Nexus instance is up",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"reason": {
						SchemaProps: spec.SchemaProps{
							Description: "Gives more information about a failure status",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"nexusRoute": {
						SchemaProps: spec.SchemaProps{
							Description: "Route for external service access",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"updateConditions": {
						VendorExtensible: spec.VendorExtensible{
							Extensions: spec.Extensions{
								"x-kubernetes-list-type": "atomic",
							},
						},
						SchemaProps: spec.SchemaProps{
							Description: "Conditions reached during an update",
							Type:        []string{"array"},
							Items: &spec.SchemaOrArray{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
					"serverOperationsStatus": {
						SchemaProps: spec.SchemaProps{
							Description: "ServerOperationsStatus describes the general status for the operations performed in the Nexus server instance",
							Ref:         ref("github.com/m88i/nexus-operator/api/v1alpha1.OperationsStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/m88i/nexus-operator/api/v1alpha1.OperationsStatus", "k8s.io/api/apps/v1.DeploymentStatus"},
	}
}
