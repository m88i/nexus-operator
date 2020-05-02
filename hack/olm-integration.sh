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

source ./hack/export-version.sh

operator-sdk generate csv --apis-dir ./pkg/apis/apps/v1alpha1 --verbose --operator-name nexus-operator --csv-version $OP_VERSION
operator-sdk bundle create quay.io/m88i/nexus-operator.v${OP_VERSION}:latest  -d ./deploy/olm-catalog/nexus-operator/${OP_VERSION} --package nexus-operator --channels alpha  --overwrite --image-builder podman
podman push quay.io/m88i/nexus-operator.v${OP_VERSION}:latest
operator-sdk bundle validate quay.io/m88i/nexus-operator.v${OP_VERSION}:latest

# For future reference when creating a cool CI:
# deployment testing: https://operator-sdk.netlify.app/docs/olm-integration/olm-deployment/
# operator-sdk olm install
# operator-sdk run --olm --operator-namespace nexus --operator-version ${VERSION}
# kubectl apply -f ./examples/nexus3-centos-no-volume.yaml
# after testing ....
# operator-sdk cleanup --olm --operator-namespace nexus --operator-version ${VERSION}
# operator-sdk olm uninstall