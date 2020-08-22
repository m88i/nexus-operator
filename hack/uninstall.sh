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


NAMESPACE=nexus

echo "....... Uninstalling ......."
echo "....... Deleting CRDs......."
kubectl delete -f deploy/crds/apps.m88i.io_nexus_crd.yaml

echo "....... Deleting Rules and Service Account ......."
kubectl delete -f deploy/role.yaml -n ${NAMESPACE}
kubectl delete -f deploy/role_binding.yaml -n ${NAMESPACE}
kubectl delete -f deploy/service_account.yaml -n ${NAMESPACE}

echo "....... Deleting Operator ......."
kubectl delete -f deploy/operator.yaml -n ${NAMESPACE}

echo "....... Deleting namespace ${NAMESPACE}......."
kubectl delete namespace ${NAMESPACE}
