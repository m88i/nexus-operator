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


# The git command will fetch the most recent tag across all branches
LATEST_TAG=$(git describe --tags $(git rev-list --tags --max-count=1))
NAMESPACE=nexus

echo "[INFO]  The repository will be checked out at the latest release"
echo "....... Checkout code at ${LATEST_TAG} ......"
git checkout tags/${LATEST_TAG}

echo "....... Creating namespace ......."
kubectl create namespace ${NAMESPACE}

echo "....... Applying CRDS ......."
kubectl apply -f deploy/crds/apps.m88i.io_nexus_crd.yaml

echo "....... Applying Rules and Service Account ......."
kubectl apply -f deploy/role.yaml -n ${NAMESPACE}
kubectl apply -f deploy/role_binding.yaml -n ${NAMESPACE}
kubectl apply -f deploy/service_account.yaml -n ${NAMESPACE}

echo "....... Applying Nexus Operator ......."
kubectl apply -f deploy/operator.yaml -n ${NAMESPACE}

echo "....... Creating the Nexus 3.x Server ......."
kubectl apply -f examples/nexus3-centos-no-volume.yaml -n ${NAMESPACE}
