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

# CUSTOM_IMAGE_TAG: name of the operator tag. set this var to change the default name of the image being built. Default to quay.io/m88i/nexus-operator:current-version
# BUILDER: builder to build the image. either podman or docker. default to podman

# include
source ./hack/go-mod-env.sh
source ./hack/export-version.sh

REGISTRY=quay.io/m88i
IMAGE=nexus-operator
TAG=${OP_VERSION}
DEFAULT_BASE_IMAGE=registry.redhat.io/ubi8/ubi-minimal:latest

setGoModEnv
go generate ./...

if [[ ! -z ${CUSTOM_BASE_IMAGE} ]]; then
    sed -i -e 's,'"${DEFAULT_BASE_IMAGE}"','"${CUSTOM_BASE_IMAGE}"',' ./build/Dockerfile
fi
if [[ -z ${CUSTOM_IMAGE_TAG} ]]; then
    CUSTOM_IMAGE_TAG=${REGISTRY}/${IMAGE}:${TAG}
fi
if [[ -z ${BUILDER} ]]; then
    BUILDER=podman
fi

# changed to podman, see: https://www.linuxuprising.com/2019/11/how-to-install-and-use-docker-on-fedora.html
operator-sdk build ${CUSTOM_IMAGE_TAG} --image-builder ${BUILDER}
