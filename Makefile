# Identify Operating System

GOLANGCI_LINT_VERSION = 1.61.0
CURRENT_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
BIN_DIR := "${CURRENT_DIR}bin"
DOCKER_IMAGE_TAG ?= get-version
export PATH+=$(BIN_DIR):

install-golangci-lint-if-needed:
	@if ! "${BIN_DIR}/golangci-lint" --version | grep '${GOLANGCI_LINT_VERSION}' >/dev/null 2>&1 ; then \
		mkdir -p "${BIN_DIR}"; \
  		echo "Installing golangci-lint to '${BIN_DIR}'"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b bin/ v$(GOLANGCI_LINT_VERSION); \
	fi

.PHONY: lint
lint: install-golangci-lint-if-needed
	@${BIN_DIR}/golangci-lint run

.PHONY: test
test:
	go test -v ./...

.PHONY: check
check: lint test

fix: install-golangci-lint-if-needed
	@${BIN_DIR}/golangci-lint run --fix

build-binary:
	@mkdir build >/dev/null 2>&1 || true
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o ./build/get-version .

build-docker-image: build-binary
	@echo 'Building docker image "${DOCKER_IMAGE_TAG}"'
	@docker build -t ${DOCKER_IMAGE_TAG} -f ./Dockerfile build/
