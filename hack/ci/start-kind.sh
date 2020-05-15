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

set -e

if [ -z "${KUBE_VERSION}" ]; then
    KIND_KUBE_VERSION=""
else
    KIND_KUBE_VERSION="--image kindest/node:${KUBE_VERSION}"
fi

if [ ! -z ${VERBOSE+x} ] && [ "${VERBOSE}" == "1" ]; then
    KIND_VERBOSITY="--verbosity 1"
else
    KIND_VERBOSITY="--verbosity 0"
fi

if [[ ! "$(./bin/kind get clusters)" =~ "${CLUSTER_NAME}" ]]; then
    ./bin/kind create cluster --name ${CLUSTER_NAME} --wait 1m ${KIND_KUBE_VERSION} ${KIND_VERBOSITY}
else
    echo "---> Already found cluster named '${CLUSTER_NAME}'"
fi

echo "---> Checking KIND cluster conditions"
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide
