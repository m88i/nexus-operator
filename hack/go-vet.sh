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

set -e
GOPATH=$(go env GOPATH)
./hack/go-mod.sh
./hack/addheaders.sh

operator-sdk generate k8s
operator-sdk generate crds

# get the openapi binary
which openapi-gen >/dev/null || go build -o $GOPATH/openapi-gen k8s.io/kube-openapi/cmd/openapi-gen
echo "Generating openapi files"
openapi-gen --logtostderr=true -o "" -i ./pkg/apis/apps/v1alpha1 -O zz_generated.openapi -p ./pkg/apis/apps/v1alpha1 -h ./hack/boilerplate.go.txt -r "-"

./hack/generate-manifests.sh

go vet ./...
