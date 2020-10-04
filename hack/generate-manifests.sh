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

source ./hack/export-version.sh


# generate manifests for package format
# use --from-version to upgrade: https://operator-sdk.netlify.app/docs/olm-integration/generating-a-csv/#upgrading-your-csv
operator-sdk generate csv --apis-dir ./pkg/apis/apps/v1alpha1 --update-crds --csv-version $OP_VERSION --make-manifests=false --verbose --operator-name nexus-operator

# manifests format for bundle image
operator-sdk generate csv --apis-dir ./pkg/apis/apps/v1alpha1 --verbose --operator-name nexus-operator --csv-version $OP_VERSION

# our package doesn't have the same name as the operator
rm ./deploy/olm-catalog/nexus-operator/nexus-operator.package.yaml -rf

source ./hack/generate-yaml-installer.sh