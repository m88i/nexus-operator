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

if [[ -n ${CUSTOM_BASE_IMAGE} ]]; then
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
