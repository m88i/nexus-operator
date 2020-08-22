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


set -e
GOPATH=$(go env GOPATH)
./hack/go-mod.sh
./hack/addheaders.sh

operator-sdk generate k8s
operator-sdk generate crds

# get the openapi binary
command -v openapi-gen >/dev/null || go build -o $GOPATH/openapi-gen k8s.io/kube-openapi/cmd/openapi-gen
echo "Generating openapi files"
openapi-gen --logtostderr=true -o "" -i ./pkg/apis/apps/v1alpha1 -O zz_generated.openapi -p ./pkg/apis/apps/v1alpha1 -h ./hack/boilerplate.go.txt -r "-"

./hack/generate-manifests.sh

go vet ./...
