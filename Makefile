.DEFAULT_GOAL: test-unit

############
# Building #
############

build-agent: ## Build agent binary
	go build -o ./bin/agent ./cmd/agent
.PHONY: build-agent

build-lpr-cli: ## Build client binary
	go build -o ./bin/lpr ./cmd/cli
.PHONY: build-agent

###########
# Testing #
###########

test-unit: ## Execute unit tests
	go test -v -race -count=1 ./...
.PHONY: test-unit

test-unit-hammer: ## Execute hammer tests in the project to spot eventual test instability
	go test -count=100 -short -failfast ./...
.PHONY: test-unit-hammer

test-cover-html: ## Generate file with unit test coverage data
	go test -v -race -count=1 -coverprofile=./coverage.txt ./...
	go tool cover -html=./coverage.txt
.PHONY: test-cover-html

test-lint:
	./hack/run-lint.sh
.PHONY: test-lint

###############
# Development #
###############

fix-lint-issues:
	LINT_FORCE_FIX=true ./hack/run-lint.sh
.PHONY: fix-lint

##############
# Generating #
##############

gen-grpc-resources: ## Generate gRPC + ProtoBuf Go code for client and server
	./hack/gen-grpc-resources.sh
.PHONY: gen-proto-source-code

gen-certs: ## Generate certificates for client and server
	./hack/gen-certs.sh
.PHONY: gen-proto-source-code

#############
# Other     #
#############

help: ## Show this help
	@egrep -h '\s##\s' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
.PHONY: help
