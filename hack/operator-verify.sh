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


source ./hack/export-version.sh

OUTPUT="${PWD}/build/_output/operatorhub"
OP_PATH="community-operators/nexus-operator"
OPERATOR_TESTING_IMAGE=quay.io/operator-framework/operator-testing:latest

echo "Output dir is set to ${OUTPUT}"

# clean up
rm -rf "${OUTPUT}"
mkdir -p "${OUTPUT}/nexus-operator/${OP_VERSION}"

cp "./deploy/olm-catalog/nexus-operator/${OP_VERSION}/"*.yaml "${OUTPUT}/nexus-operator/${OP_VERSION}"
cp ./deploy/olm-catalog/nexus-operator/nexus-operator.package.yaml "${OUTPUT}/nexus-operator"

docker pull ${OPERATOR_TESTING_IMAGE}
docker run --rm -v ${OUTPUT}:/community-operators:z ${OPERATOR_TESTING_IMAGE} operator.verify --no-print-directory OP_PATH=${OP_PATH} VERBOSE=true
