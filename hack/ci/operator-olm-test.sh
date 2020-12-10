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


source ./hack/verify-version.sh
source ./hack/ci/operator-ensure-manifest.sh

if [ -z ${KUBECONFIG} ]; then
    KUBECONFIG=${HOME}/.kube/config
    echo "---> KUBECONFIG environment variable not set, defining to:"
    ls -la ${KUBECONFIG}
fi

csv_file=${OUTPUT}/nexus-operator-m88i/${OP_VERSION}/nexus-operator.v${OP_VERSION}.clusterserviceversion.yaml
echo "---> Updating CSV file '${csv_file}' to imagePullPolicy: Never"
sed -i 's/imagePullPolicy: Always/imagePullPolicy: Never/g' ${csv_file}
echo "---> Resulting imagePullPolicy on manifest files"
grep -rn imagePullPolicy ${OUTPUT}/nexus-operator-m88i


bash <(curl -sL https://cutt.ly/operator-test) \
all \
${output}/nexus-operator-m88i/${OP_VERSION}
