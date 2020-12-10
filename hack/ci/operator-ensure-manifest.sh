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

OUTPUT="${PWD}/build/_output/operatorhub/upstream-community-operators"

echo "---> Output dir is set to ${OUTPUT}"

# clean up
rm -rf "${OUTPUT}"

mkdir -p "${OUTPUT}"

rm -rf ~/operators/
git clone --depth 1 --filter=blob:none https://github.com/operator-framework/community-operators.git ~/operators/

cp -r ~/operators/community-operators/nexus-operator-m88i "${OUTPUT}"
rm -rf ~/operators/

mkdir -p "${OUTPUT}/nexus-operator-m88i/${OP_VERSION}"
cp -v "./bundle/manifests/apps.m88i.io_nexus.yaml" "${OUTPUT}/nexus-operator-m88i/${OP_VERSION}/apps.m88i.io_nexus_crd.yaml"
cp -v "./bundle/manifests/nexus-operator.clusterserviceversion.yaml" "${OUTPUT}/nexus-operator-m88i/${OP_VERSION}/nexus-operator.v${OP_VERSION}.clusterserviceversion.yaml"
cp -v "./bundle/nexus-operator-m88i.package.yaml" "${OUTPUT}/nexus-operator-m88i/"

sed -i "s/{version}/${OP_VERSION}/g" "${OUTPUT}/nexus-operator-m88i/nexus-operator-m88i.package.yaml"

echo "---> Manifest files in the output directory for OLM verification"
ls -la "${OUTPUT}/nexus-operator-m88i/"
