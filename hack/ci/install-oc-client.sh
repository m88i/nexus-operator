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


if [ ! -f ./bin/oc ]; then
    echo "Installing Openshift CLI"
    curl -LO https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz
    tar -xvf openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz
    mv openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit/oc .
    chmod +x ${PWD}/oc &&
        mkdir -p ./bin &&
        mv ${PWD}/oc ./bin/oc &&
        rm ${PWD}/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit* -rf
else
    echo "OpenShift CLI already installed, skipping"
fi
