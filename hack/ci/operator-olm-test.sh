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

source ./hack/ci/operator-ensure-manifest.sh

CATALOG_IMAGE="operatorhubio-catalog:temp"
OP_PATH="community-operators/nexus-operator"
INSTALL_MODE="SingleNamespace"
OPERATOR_TESTING_IMAGE=quay.io/operator-framework/operator-testing:latest

if [ -z ${KUBECONFIG} ]; then
    KUBECONFIG=${HOME}/.kube/config
    echo "---> KUBECONFIG environment variable not set, defining to:"
    ls -la ${KUBECONFIG}
fi

echo "---> Building temporary catalog Image"
docker build --build-arg PERMISSIVE_LOAD=false -f ./hack/ci/operatorhubio-catalog.Dockerfile -t ${CATALOG_IMAGE} .
echo "---> Loading Catalog Image into Kind"
kind load docker-image ${CATALOG_IMAGE} --name operator-test

# running tests
docker pull ${OPERATOR_TESTING_IMAGE}
docker run --network=host --rm \
    -v ${KUBECONFIG}:/root/.kube/config:z \
    -v ${OUTPUT}:/community-operators:z ${OPERATOR_TESTING_IMAGE} \
    operator.test --no-print-directory \
    OP_PATH=${OP_PATH} VERBOSE=true NO_KIND=0 CATALOG_IMAGE=${CATALOG_IMAGE} INSTALL_MODE=${INSTALL_MODE}
