SHELL := /bin/bash
PROJECT_PATH=/go/src/github.com/makkalot/eskit
REPOSITORY_NAME=makkalot
NAMESPACE= default
GCLOUD_CONTEXT=gke_verdant-tempest-186207_europe-west1-d_simple-micro-cluster

PROJECT_NAME=eskit
HOSTNAME_SUFFIX=$(PROJECT_NAME).makkalot.com

CI_JOB_ID?=eskit
COMPOSE=docker compose -p $(CI_JOB_ID)
COMMON_SH=source ./scripts/common.sh &&
GINKGO_FOCUS?=integration

HELM_IMAGE_NAME=$(REPOSITORY_NAME)/eskit-helm:latest

KUBECONFIG_DIR=/tmp/kube
KUBECONFIG_PATH=/tmp/kube/config

COMPOSE_FILE_UNIT=docker-compose-unit.yml
COMPOSE_FILE_INTEGRATION=docker-compose.yml

ifndef CI
	KUBECONFIG_DIR=/tmp/kube
	KUBECONFIG_PATH=/tmp/kube/config

	HELM_SH=helm
	KUBECTL_SH=kubectl
else
	KUBECONFIG_DIR=`pwd`/.kube
	KUBECONFIG_PATH=$(KUBECONFIG_DIR)/config

	CONTAINER_SH=docker run -v `pwd`:/go/src -w /go/src -e KUBECONFIG=/go/src/.kube/config --rm $(HELM_IMAGE_NAME)
	HELM_SH=$(CONTAINER_SH) helm
	KUBECTL_SH=$(CONTAINER_SH) kubectl

endif

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	BASE64_DECODE=base64 -D
else
	BASE64_DECODE=base64 -d
endif

DEPLOY_TARGETS =
K8S_SERVICES =

GO_BUILD_SERVICES =
GO_BUILD_SERVICES += users/cmd/users
GO_BUILD_SERVICES += camconfig/cmd/camconfig
GO_BUILD_SERVICES += admin/cmd/admin
GO_BUILD_TARGETS = $(addprefix ./.bin/, $(GO_BUILD_SERVICES))


.PHONY: all
all: build

.PHONY: build
build: build-compose-go

# Cleaning UP, every section has a clean-{section}-{cm1}.. which tries to be reverse
# of what's beeen done. For example clean-generate cleans all the generated files
# And glean-generate-grpc cleans up all the generated grpc files

.PHONY: clean
clean: clean-build

# Dependency targets

.PHONY: deps
deps: deps-go

.PHONY: deps-go
deps-go:
	go get -v -u ./...
	go mod tidy


# Build related tasks

.PHONY: build-compose-go
build-compose-go:
	$(COMMON_SH) source ./scripts/compose.sh && compose_build $(CI_JOB_ID)

.PHONY: build-go
build-go: clean-build-go $(GO_BUILD_TARGETS)

$(GO_BUILD_TARGETS): build-deps-go
	mkdir -p ./bin
	SERVICE_NAME=$(shell basename $@) && \
	SERVICE_PATH=$(shell echo $@ | cut -c6-) && \
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o ./bin/$$SERVICE_NAME ./services/$$SERVICE_PATH

.PHONY: build-deps-go
build-deps-go: $(wildcard services/**/*.go)


.PHONY: clean-build-go
clean-build-go:
	for build_target in $(GO_BUILD_SERVICES); do \
		rm -rf ./bin ; \
	done

.PHONY: clean-build
clean-build: clean-build-go


# Deployment related targets

.PHONY: deploy-compose
deploy-compose: clean-build build-compose-go
	$(COMMON_SH) source ./scripts/compose.sh && compose_deploy $(CI_JOB_ID)

# TEST targets will be here

.PHONY: test
# test: build test-compose-go test-compose-py
test: build test-compose-go

.PHONY: test-compose-go
test-compose-go: test-compose-go-unit test-compose-go-integration

.PHONY: test-compose-go-unit
test-compose-go-unit:
	export COMPOSE_FILE=$(COMPOSE_FILE_UNIT); \
	$(COMMON_SH) source ./scripts/compose.sh && compose_tests_golang $(CI_JOB_ID)

.PHONY: test-compose-go-integration
test-compose-go-integration: build-compose-go
	export COMPOSE_FILE=$(COMPOSE_FILE_INTEGRATION); \
	$(COMMON_SH) source ./scripts/compose.sh && compose_tests_golang $(CI_JOB_ID)

.PHONY: test-go
test-go: test-go-unit test-go-integration

.PHONY: test-go-unit
test-go-unit:
	cd lib && go test -count=1 -v ./...

.PHONY: test-go-integration
test-go-integration:
	cd tests && ginkgo -r -v

# Generate targets removed - no longer using gRPC/protobuf