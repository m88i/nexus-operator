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

# make the script exit if any command fails
set -e

UNABLE_TO_PUSH_ERR="Unable to push. You'll need to push manually."

confirm_git_or_die () {
    if ! command -v git >/dev/null 2>&1; then
        echo "Could not find git. You'll need to push manually." >&2
        exit
    fi
}

push_with_defaults_or_die () {
    confirm_git_or_die
    curr_branch=$(git rev-parse --abbrev-ref HEAD)
    echo "Pushing to origin/${curr_branch}"
    git push -u origin "${curr_branch}" || echo "${UNABLE_TO_PUSH_ERR}" >&2; exit 1
}

prompt_and_push_or_die () {
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
    git push -u "${remote}" "${branch}" || echo "${UNABLE_TO_PUSH_ERR}" >&2; exit 1
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
        [Yy]* ) prompt_and_push_or_die; exit 0;;
        [Nn]* ) exit 0;;
        * ) echo "Please answer y or n.";;
    esac
done