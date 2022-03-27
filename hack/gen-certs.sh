#!/usr/bin/env bash

# standard bash error handling
set -o nounset # treat unset variables as an error and exit immediately.
set -o errexit # exit immediately when a command fails.
set -E         # needs to be set if we want the ERR trap

CURRENT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
REPO_ROOT_DIR=$(cd "${CURRENT_DIR}/.." && pwd)
readonly CURRENT_DIR
readonly REPO_ROOT_DIR
readonly CERT_DIR="${REPO_ROOT_DIR}/certs"

# shellcheck source=./hack/lib/utilities.sh
source "${CURRENT_DIR}/lib/utilities.sh" || { echo 'Cannot load CI utilities.' exit 1; }

cleanup() {
	find "${CERT_DIR}" -name '*_csr.pem' -delete
}

trap cleanup EXIT

gen::agent::ca() {
	shout "Generate Agent CA"
	openssl req -x509 \
		-newkey rsa:4096 \
		-nodes \
		-days 3650 \
		-keyout agent_ca_key.pem \
		-out agent_ca_cert.pem \
		-subj /C=US/ST=CA/L=SVL/O=gRPC/CN=agent_ca/ -config ./openssl.cnf \
		-extensions test_ca \
		-sha256
}

gen::agent::cert() {
	shout "Generate Agent certificates"
	openssl genrsa -out agent_key.pem 4096
	openssl req -new \
		-key agent_key.pem \
		-days 3650 \
		-out agent_csr.pem \
		-subj /C=US/ST=CA/L=SVL/O=gRPC/CN=agent/ \
		-config ./openssl.cnf \
		-reqexts test_agent
	openssl x509 -req \
		-in agent_csr.pem \
		-CAkey agent_ca_key.pem \
		-CA agent_ca_cert.pem \
		-days 3650 \
		-set_serial 1000 \
		-out agent_cert.pem \
		-extfile ./openssl.cnf \
		-extensions test_agent \
		-sha256

	openssl verify -verbose -CAfile agent_ca_cert.pem agent_cert.pem
}

gen::client::ca() {
	shout "Generate Client CA"
	openssl req -x509 \
		-newkey rsa:4096 \
		-nodes \
		-days 3650 \
		-keyout client_ca_key.pem \
		-out client_ca_cert.pem \
		-subj /C=US/ST=CA/L=SVL/O=gRPC/CN=client_ca/ \
		-config ./openssl.cnf \
		-extensions test_ca \
		-sha256
}

# Required envs:
#  - USER_NAME
#  - USER_ROLE
#
# usage: env USER_NAME=Ricky USER_ROLE=admin gen::client::cert
gen::client::cert() {
	shout "Generate Client certificates for ${USER_NAME} with ${USER_ROLE} role"

	openssl genrsa -out "${USER_NAME}_client_key.pem" 4096
	openssl req -new \
		-key "${USER_NAME}_client_key.pem" \
		-days 3650 \
		-out "${USER_NAME}_client_csr.pem" \
		-subj /C=US/ST=CA/L=SVL/O="${USER_ROLE}"/CN="${USER_NAME}"/ \
		-config ./openssl.cnf \
		-reqexts test_client
	openssl x509 -req \
		-in "${USER_NAME}_client_csr.pem" \
		-CAkey client_ca_key.pem \
		-CA client_ca_cert.pem \
		-days 3650 \
		-set_serial 1000 \
		-out "${USER_NAME}_client_cert.pem" \
		-extfile ./openssl.cnf \
		-extensions test_client \
		-sha256

	openssl verify -verbose -CAfile client_ca_cert.pem "${USER_NAME}_client_cert.pem"
}

main() {
	pushd "${CERT_DIR}" >/dev/null

	rm -f *.pem
	gen::agent::ca
	gen::agent::cert

	gen::client::ca

	USER_NAME="Ricky"
	USER_ROLE="admin"
	gen::client::cert

	USER_NAME="Groot"
	USER_ROLE="user"
	gen::client::cert

	popd >/dev/null
}

main
