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


source ./hack/export-version.sh

operator-sdk generate csv --apis-dir ./pkg/apis/apps/v1alpha1 --verbose --operator-name nexus-operator --csv-version $OP_VERSION
operator-sdk bundle create quay.io/m88i/nexus-operator.v${OP_VERSION}:latest -d ./deploy/olm-catalog/nexus-operator/${OP_VERSION} --package nexus-operator-m88i --channels alpha --overwrite --image-builder podman
podman push quay.io/m88i/nexus-operator.v${OP_VERSION}:latest
operator-sdk bundle validate quay.io/m88i/nexus-operator.v${OP_VERSION}:latest

# For future reference when creating a cool CI:
# deployment testing: https://operator-sdk.netlify.app/docs/olm-integration/olm-deployment/
# operator-sdk olm install
# operator-sdk run --olm --operator-namespace nexus --operator-version ${VERSION}
# kubectl apply -f ./examples/nexus3-centos-no-volume.yaml
# after testing ....
# operator-sdk cleanup --olm --operator-namespace nexus --operator-version ${VERSION}
# operator-sdk olm uninstall
