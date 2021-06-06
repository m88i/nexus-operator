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

source ./hack/go-path.sh

# it's expected to have GOPATH in your PATH
# The command in or will fetch the latest tag available for golangci-lint and install in $GOPATH/bin/
command -v golangci-lint >/dev/null || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "${GOPATH}"/bin

golangci-lint run --enable=revive --skip-files=pkg/test/client.go,pkg/framework/kind/kinds.go,controllers/nexus/resource/validation/defaults.go
