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


# make the script exit if any command fails
set -e

UNABLE_TO_PUSH_ERR="Unable to push. You'll need to push manually."

confirm_git_or_die() {
    if ! command -v git >/dev/null 2>&1; then
        echo "Could not find git. You'll need to push manually." >&2
        exit
    fi
}

push_with_defaults_or_die() {
    confirm_git_or_die
    curr_branch=$(git rev-parse --abbrev-ref HEAD)
    echo "Pushing to origin/${curr_branch}"
    git push origin "${curr_branch}" || (
        echo "${UNABLE_TO_PUSH_ERR}" >&2
        exit 1
    )
}

prompt_and_push_or_die() {
    confirm_git_or_die
    read -rp "Insert the remote name: [origin] " remote
    if [[ $remote == "" ]]; then
        remote="origin"
    fi

    curr_branch=$(git rev-parse --abbrev-ref HEAD)
    read -rp "Insert branch: [${curr_branch}] " branch
    if [[ $branch == "" ]]; then
        branch=$curr_branch
    fi

    echo "Pushing to ${remote}/${branch}"
    git push "${remote}" "${branch}" || (
        echo "${UNABLE_TO_PUSH_ERR}" >&2
        exit 1
    )
}

./hack/go-test.sh
./hack/go-build.sh
./hack/run-e2e-test.sh

echo "All tests were successful!"

if [[ ${PUSH_WITH_DEFAULTS^^} == "TRUE" ]]; then
    push_with_defaults_or_die
    exit 0
fi

while true; do
    read -rp "Do you wish to push? (y/n) " yn
    case $yn in
    [Yy]*)
        prompt_and_push_or_die
        exit 0
        ;;
    [Nn]*) exit 0 ;;
    *) echo "Please answer y or n." ;;
    esac
done
