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

# it's expected to have GOPATH in your PATH
# see this link to understand why we are using a different go.mod file: https://github.com/golang/go/issues/34506
which golangci-lint >/dev/null || go get -modfile=go.tools.mod github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0

golangci-lint run ./pkg/... ./test/...
