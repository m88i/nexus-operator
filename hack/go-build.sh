#!/bin/sh
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


. ./hack/go-mod-env.sh

REPO=https://github.com/m88i/nexus-operator
BRANCH=master
REGISTRY=quay.io/m88i
IMAGE=nexus-operator
TAG=0.2.0
TAR=${BRANCH}.tar.gz
URL=${REPO}/archive/${TAR}
CFLAGS=""

setGoModEnv
go generate ./...
if [[ -z ${CI} ]]; then
    ./hack/go-test.sh
    operator-sdk build ${REGISTRY}/${IMAGE}:${TAG}
else
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -v -a -o build/_output/bin/nexus-operator github.com/m88i/nexus-operator/cmd/manager
fi
