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

default_kind_version=v0.8.1

if [[ -z ${KIND_VERSION} ]]; then
    KIND_VERSION=$default_kind_version
fi

GOPATH=$(go env GOPATH)

if [[ $(which kind) ]]; then
  echo "---> kind is already installed. Please make sure it is the required ${KIND_VERSION} version before proceeding"
else
  echo "---> kind not found, installing it in \$GOPATH/bin/"
  curl -L https://kind.sigs.k8s.io/dl/$KIND_VERSION/kind-$(uname)-amd64 -o "$GOPATH"/bin/kind
  chmod +x "$GOPATH"/bin/kind
fi

#for verification
kind version