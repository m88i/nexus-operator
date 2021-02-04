#!/bin/bash
# Copyright 2020 Nexus Operator and/or its authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

command -v yq > /dev/null || ( echo "Please install yq before proceeding (https://pypi.org/project/yq/)" && exit 1 )
command -v kustomize > /dev/null || go build -o "${GOPATH}"/bin/kustomize sigs.k8s.io/kustomize/kustomize/v3@v3.5.4 || exit 1

# first, let's filter out all manifests for kinds we don't care about
# 'select(.kind != "ValidatingWebhookConfiguration" and .kind != "Issuer" and .kind != "Certificate" and .kind != "MutatingWebhookConfiguration" and .metadata.name != "nexus-operator-webhook-service")'

# then delete the volumes which would contain the certs
# 'del(.. | .volumes?, .volumeMounts?)'

# then finally insert the env var which disables webhooks
# 'if .kind=="Deployment" then .spec.template.spec.containers[1].env[0]={"name":"USE_WEBHOOKS", "value":"FALSE"} else . end'

kustomize build config/default/ | yq -Y 'select(.kind != "ValidatingWebhookConfiguration" and .kind != "Issuer" and .kind != "Certificate" and .kind != "MutatingWebhookConfiguration" and .metadata.name != "nexus-operator-webhook-service")' \
 | yq -Y 'del(.. | .volumes?, .volumeMounts?)' \
 | yq -Y 'if .kind=="Deployment" then .spec.template.spec.containers[1].env[0]={"name":"USE_WEBHOOKS", "value":"FALSE"} else . end' > webhookless-nexus-operator.yaml
