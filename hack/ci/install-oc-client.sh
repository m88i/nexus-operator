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
