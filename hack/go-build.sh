#!/bin/sh

. ./hack/go-mod-env.sh

REPO=https://github.com/m88i/nexus-operator
BRANCH=master
REGISTRY=quay.io/m88i
IMAGE=nexus-operator
TAG=0.1.0
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
