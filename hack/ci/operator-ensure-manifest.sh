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

echo "---> Output dir is set to ${OUTPUT}"

# clean up
rm -rf "${OUTPUT}"

mkdir -p "${OUTPUT}"
cp -r "./deploy/olm-catalog/nexus-operator/" "${OUTPUT}/"
mv "${OUTPUT}/nexus-operator" "${OUTPUT}/nexus-operator-m88i"
rm "${OUTPUT}/nexus-operator-m88i/manifests" -rf
rm "${OUTPUT}/nexus-operator-m88i/metadata" -rf
rm "${OUTPUT}/nexus-operator-m88i/0.1.0" -rf

echo "---> Manifest files in the output directory for OLM verification"
ls -la "${OUTPUT}/nexus-operator-m88i/"
