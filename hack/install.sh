#!/bin/bash
#     Copyright 2019 Nexus Operator and/or its authors
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


NAMESPACE=nexus

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
kubectl apply -f deploy/examples/nexus3-centos.yaml -n ${NAMESPACE}
