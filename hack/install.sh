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


VERSION=$1

if [ -z ${VERSION} ]; then
    echo "Please inform the desired version"
    exit(1)
fi

echo "Downloading latest version"
curl -LO https://github.com/m88i/nexus-operator/releases/download/${VERSION}/nexus-operator.yaml

echo "....... Installing Nexus Operator ......."

kubectl apply -f nexus-operator.yaml