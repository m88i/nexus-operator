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


if [[ -z ${NAMESPACE_E2E} ]]; then
    NAMESPACE_E2E="nexus-e2e"
fi

if [[ -z ${TIMEOUT_E2E} ]]; then
    TIMEOUT_E2E="15m"
fi

if [[ ${CREATE_NAMESPACE^^} == "TRUE" ]]; then
    echo "---> Creating Namespace ${NAMESPACE_E2E} to run e2e tests"
    kubectl create namespace $NAMESPACE_E2E
else
    echo "---> Skipping creating namespace"
fi

echo "---> Executing e2e tests on ${NAMESPACE_E2E}"

if [[ ${RUN_WITH_IMAGE^^} == "TRUE" ]]; then
    echo "---> Running tests with image instead of local"
    # see: https://kind.sigs.k8s.io/docs/user/quick-start/#loading-an-image-into-your-cluster
    echo "---> Updating deployment file to imagePullPolicy: Never"
    sed -i.bak 's/imagePullPolicy:\s*Always/imagePullPolicy: Never/g' ./deploy/operator.yaml
    echo "---> Updating cluster role binding to SA on ${NAMESPACE_E2E}"
    sed -i.bak "s/\${NAMESPACE}/${NAMESPACE_E2E}/g" ./deploy/role_binding.yaml
    
    kubectl delete -f ./deploy/service_account.yaml
    kubectl delete -f ./deploy/role_binding.yaml
    kubectl delete -f ./deploy/role.yaml
    kubectl delete -f ./deploy/operator.yaml

    operator-sdk test local ./test/e2e --go-test-flags "-v -timeout $TIMEOUT_E2E $ADDITIONAL_FLAGS" --debug --operator-namespace $NAMESPACE_E2E
    test_exit_code=$?
    mv -f ./deploy/operator.yaml.bak ./deploy/operator.yaml
    mv -f ./deploy/role_binding.yaml.bak ./deploy/role_binding.yaml
else
    echo "---> Running tests with local binary"
    operator-sdk test local ./test/e2e --go-test-flags "-v -timeout $TIMEOUT_E2E $ADDITIONAL_FLAGS" --debug --up-local --operator-namespace $NAMESPACE_E2E
    test_exit_code=$?
fi

if [[ ${CREATE_NAMESPACE^^} == "TRUE" ]]; then
    echo "---> Cleaning up namespace ${NAMESPACE_E2E}"
    kubectl delete namespace $NAMESPACE_E2E
fi

if [ $test_exit_code -eq 0 ]; then
    echo "Success: e2e test ended successfully!"
    exit 0
else
    echo "Failure: e2e test failed to run. See the logs for more information" >&2
    exit 1
fi
