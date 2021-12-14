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

# shellcheck disable=SC2155
declare readme_changed=$(git status -s | grep  'README.md' || :)
if [ -z "${readme_changed}" ]; then
  exit 0
fi

# enforce bin directory
mkdir bin

command -v bin/gh-md-toc > /dev/null || curl https://raw.githubusercontent.com/ekalinin/github-markdown-toc/master/gh-md-toc -o bin/gh-md-toc && chmod +x bin/gh-md-toc

bin/gh-md-toc --no-backup README.md
sed -i'' -e '/^<!-- Added by:/d' README.md
