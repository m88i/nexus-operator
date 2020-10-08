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
    VERSION=$(curl https://api.github.com/repos/m88i/nexus-operator/releases/latest | python -c "import sys, json; print(json.load(sys.stdin)['tag_name'])")
fi

echo "....... Installing Nexus Operator ${VERSION} ......."

kubectl apply -f https://github.com/m88i/nexus-operator/releases/download/${VERSION}/nexus-operator.yaml