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

rm ./deploy/olm-catalog/nexus-operator/manifests -rf
# generate manifests use --from-version to upgrade: https://operator-sdk.netlify.app/docs/olm-integration/generating-a-csv/#upgrading-your-csv
operator-sdk generate csv --apis-dir ./pkg/apis/apps/v1alpha1 --update-crds --csv-version $OP_VERSION --make-manifests=false --verbose --operator-name nexus-operator
operator-sdk generate csv --apis-dir ./pkg/apis/apps/v1alpha1 --verbose --operator-name nexus-operator --csv-version $OP_VERSION
