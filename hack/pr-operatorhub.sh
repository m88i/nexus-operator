#!/bin/sh
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

# prepare the package for the operatorhub.io to push our changes there
# 0. make sure that the operator is ok (e2e)
# 1. run this script
# 2. push the results of build/_output/operatorhub/ to https://github.com/operator-framework/community-operators/tree/master/community-operators/kogito-cloud-operator

version=$1
output="build/_output/operatorhub/"

if [ -z "$version" ]; then
    echo "Please inform the release version. Use X.X.X"
    exit 1
fi

if ! hash operator-courier 2>/dev/null; then
  pip3 install operator-courier
fi

# clean up
rm -rf "${output}*"
mkdir -p ${output}

# will run unit tests and generate the ultimate source for the OLM catalog
make test

# copy the generated files
cp "deploy/olm-catalog/nexus-operator/${version}/"*.yaml $output
cp deploy/olm-catalog/nexus-operator/nexus-operator.package.yaml $output

# basic verification
operator-courier verify --ui_validate_io $output

# now it's your turn to push the application and the image to quay:
# https://github.com/operator-framework/community-operators/blob/master/docs/testing-operators.md#push-to-quayio

# then push it to operatorhub
# https://github.com/operator-framework/community-operators/pulls
