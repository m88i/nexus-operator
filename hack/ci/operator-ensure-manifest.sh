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

OUTPUT="${PWD}/build/_output/operatorhub"

echo "Output dir is set to ${OUTPUT}"

# clean up
rm -rf "${OUTPUT}"

mkdir -p "${OUTPUT}/nexus-operator/${OP_VERSION}"
cp "./deploy/olm-catalog/nexus-operator/${OP_VERSION}/"*.yaml "${OUTPUT}/nexus-operator/${OP_VERSION}"
cp ./deploy/olm-catalog/nexus-operator/nexus-operator.package.yaml "${OUTPUT}/nexus-operator"

# replaces
replace_version=$(grep replaces "./deploy/olm-catalog/nexus-operator/${OP_VERSION}/nexus-operator.v${OP_VERSION}.clusterserviceversion.yaml" | cut -f2 -d'v')
if [ ! -z "${replace_version}" ]; then
    echo "Found replaces version in the new CSV: ${replace_version}. Including in the package."
    mkdir -p "${OUTPUT}/nexus-operator/${replace_version}"
    cp "./deploy/olm-catalog/nexus-operator/${replace_version}/"*.yaml "${OUTPUT}/nexus-operator/${replace_version}"
fi
