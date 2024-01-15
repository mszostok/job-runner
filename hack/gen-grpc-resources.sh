#!/usr/bin/env bash

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
TMP_DIR=$(mktemp -d)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR
readonly TMP_DIR

readonly STABLE_PROTOC_VERSION=v3.19.4
readonly STABLE_PROTOC_GEN_GOGO_VERSION=v1.3.2
readonly STABLE_PROTOC_GEN_GO_GRPC_VERSION=v1.2.0
# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.' exit 1; }

cleanup() {
	rm -rf "${TMP_DIR}"
}

trap cleanup EXIT

host::install::protoc() {
	shout "Install the protoc ${STABLE_PROTOC_VERSION} locally to a tempdir..."
	readonly TMP_BIN="${TMP_DIR}/bin"
	mkdir -p "$TMP_BIN"
	pushd "$TMP_DIR" >/dev/null

	export GOBIN="$TMP_BIN"
	export PATH="${TMP_BIN}:${PATH}"

	readonly os=$(host::os)
	readonly arch=$(host::arch)
	readonly version_without_v=${STABLE_PROTOC_VERSION#"v"}
	readonly name="protoc-${version_without_v}-${os}-${arch}"

	# download the release
	curl -L -O "https://github.com/protocolbuffers/protobuf/releases/download/${STABLE_PROTOC_VERSION}/${name}.zip"

	# extract the archive
	unzip "${name}".zip >/dev/null
	echo -e "${GREEN}√ install protoc${NC}"

	# Go plugins
	go install "github.com/gogo/protobuf/protoc-gen-gogo@${STABLE_PROTOC_GEN_GOGO_VERSION}"
	echo -e "${GREEN}√ install protoc-gen-gogo${NC}"
	go install "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${STABLE_PROTOC_GEN_GO_GRPC_VERSION}"
	echo -e "${GREEN}√ install protoc-gen-go-grpc${NC}"

	popd >/dev/null
}

main() {
	if [[ "${INSTALL_DEPS:-x}" == "true" ]]; then
		host::install::protoc
	fi

	shout "Generating gRPC related resources..."

	protoc -I="${REPO_ROOT_DIR}/proto/" \
		-I="$GOPATH/src" \
		--gogo_out="Mgoogle/protobuf/duration.proto=github.com/gogo/protobuf/types:." \
		--go-grpc_out="." \
		"${REPO_ROOT_DIR}/proto/job_runner.proto"

	shout "Generation completed successfully."
}

main
