#!/bin/bash
#     Copyright 2020 Nexus Operator and/or its authors
#
#     This file is part of Nexus Operator.
#
#     Nexus Operator is free software: you can redistribute it and/or modify
#     it under the terms of the GNU General Public License as published by
#     the Free Software Foundation, either version 3 of the License, or
#     (at your option) any later version.
#
#     Nexus Operator is distributed in the hope that it will be useful,
#     but WITHOUT ANY WARRANTY; without even the implied warranty of
#     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
#     GNU General Public License for more details.
#
#     You should have received a copy of the GNU General Public License
#     along with Nexus Operator.  If not, see <https://www.gnu.org/licenses/>.

if [[ -z ${NAMESPACE_E2E} ]]; then
    NAMESPACE_E2E="nexus-e2e"
fi

if [[ ${CREATE_NAMESPACE^^} == "TRUE" ]]; then
    echo "---> Creating Namespace ${NAMESPACE_E2E} to run e2e tests"
    kubectl create namespace $NAMESPACE_E2E
else
    echo "---> Skipping creating namespace"
fi

echo "---> Executing e2e tests on ${NAMESPACE_E2E}"

if [[ ${RUN_WITH_IMAGE^^} == "TRUE" ]]; then
    echo "---> Running tests with image ${CUSTOM_IMAGE_TAG}"
    echo "---> Updating deployment file to imagePullPolicy: Never"
    sed -i 's/imagePullPolicy: Always/imagePullPolicy: Never/g' ./deploy/operator.yaml
    operator-sdk test local ./test/e2e --go-test-flags "-v" --debug --image ${CUSTOM_IMAGE_TAG} --operator-namespace $NAMESPACE_E2E
else
    echo "---> Running tests with local binary"
    operator-sdk test local ./test/e2e --go-test-flags "-v" --debug --up-local --operator-namespace $NAMESPACE_E2E
fi

test_exit_code=$?

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
