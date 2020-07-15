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

#Make sure docker is installed before proceeding
command -v docker > /dev/null || (echo "docker is not installed. Please install it before proceeding exiting...." && exit 1)

#Make sure kubectl is installed before proceeding
command -v kubectl > /dev/null || (echo "kubectl is not installed. Please install it before proceeding exiting...." && exit 1)

#make sure kind is installed before proceeding
command -v kind > /dev/null || (echo "kind is not installed. Please install it before proceeding exiting...." && exit 1)

default_cluster_name="operator-test"

if [[ -z ${CLUSTER_NAME} ]]; then
    CLUSTER_NAME=$default_cluster_name
fi
if [[ $(kind get clusters | grep ${CLUSTER_NAME}) ]]; then
  echo "---> Cluster ${CLUSTER_NAME} already present"
else
  echo "---> Provisioning new cluster"
  kind create cluster  --name ${CLUSTER_NAME} --wait 1m
fi

echo "---> Checking KIND cluster conditions"
kubectl get nodes -o wide
kubectl get pods --all-namespaces -o wide