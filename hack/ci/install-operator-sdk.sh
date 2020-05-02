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

if [[ -z ${OPERATOR_SDK_VERSION} ]]; then
    OPERATOR_SDK_VERSION=0.17.0
fi

if [ ! -f ./bin/operator-sdk ]; then
    echo "Installing Operator SDK version ${OPERATOR_SDK_VERSION}"
    curl -LO https://github.com/operator-framework/operator-sdk/releases/download/v${OPERATOR_SDK_VERSION}/operator-sdk-v${OPERATOR_SDK_VERSION}-x86_64-linux-gnu

    chmod +x ${PWD}/operator-sdk-v${OPERATOR_SDK_VERSION}-x86_64-linux-gnu &&
        mkdir -p ./bin &&
        cp ${PWD}/operator-sdk-v${OPERATOR_SDK_VERSION}-x86_64-linux-gnu ./bin/operator-sdk &&
        rm ${PWD}/operator-sdk-v${OPERATOR_SDK_VERSION}-x86_64-linux-gnu
else
    echo "Operator already installed, skipping"
fi
