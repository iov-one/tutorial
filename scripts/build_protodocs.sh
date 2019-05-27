#!/bin/bash
set -o errexit -o nounset -o pipefail
# command -v shellcheck > /dev/null && shellcheck "$0"

# This is a version of build_protodocs, which uses docker images to build, and includes gogoproto support,
# so it doesn't need the pre-clean step

out=${1:-docs/proto}

user=$(id -u):$(id -g)
protoc="docker run --rm --user ${user} --mount type=bind,source=$(pwd),target=/work --tmpfs /tmp:exec iov1/prototool-docker protoc"
prototool="docker run --rm --user ${user} --mount type=bind,source=$(pwd),target=/work --tmpfs /tmp:exec iov1/prototool-docker prototool"

mkdir -p ${out}
files=$(${prototool} files | grep -v examples | grep -v cmd/bcpd | sort)
${protoc} -I . -I /usr/include --doc_out=${out} --doc_opt=html,index.html ${files}