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
default_kind_version="v0.8.1"

if [[ -z ${KIND_VERSION+x} ]]; then
    KIND_VERSION=$default_kind_version
    echo "Using default KIND version ${KIND_VERSION}"
fi

if [ ! -f ./bin/kind ]; then
    echo "Installing KIND"
    curl -Lo ./kind https://github.com/kubernetes-sigs/kind/releases/download/${KIND_VERSION}/kind-linux-amd64
    chmod +x ${PWD}/kind &&
        mkdir -p ./bin &&
        mv ${PWD}/kind ./bin/kind
else
    echo "KIND already installed, skipping"
fi
