SHELL := /bin/bash
PROJECT_PATH=/go/src/github.com/makkalot/eskit
GRPC_IMAGE_NAME = grpc-micro:latest
SWAGGER_IMAGE_NAME=quay.io/goswagger/swagger:0.10.0
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

GRPC_CONTRACTS =
GRPC_CONTRACTS += common
GRPC_CONTRACTS += eventstore
GRPC_CONTRACTS += crudstore
GRPC_CONTRACTS += consumerstore
GRPC_CONTRACTS += users
GRPC_TARGETS = $(addprefix generated/grpc/go/, $(GRPC_CONTRACTS))

GRPC_PYTHON_CONTRACTS =
GRPC_PYTHON_CONTRACTS += common
GRPC_PYTHON_CONTRACTS += eventstore
GRPC_PYTHON_CONTRACTS += crudstore
GRPC_PYTHON_CONTRACTS += consumerstore
GRPC_PYTHON_CONTRACTS += users
GRPC_PYTHON_TARGETS = $(addprefix pyservices/generated/, $(GRPC_PYTHON_CONTRACTS))

DEPLOY_TARGETS =
K8S_SERVICES =


GO_BUILD_SERVICES =
GO_BUILD_SERVICES += users/cmd/users
GO_BUILD_SERVICES += users/cmd/usersgw
GO_BUILD_TARGETS = $(addprefix ./.bin/, $(GO_BUILD_SERVICES))


.PHONY: all
all: build

.PHONY: build
build: build-compose-go

# Cleaning UP, every section has a clean-{section}-{cm1}.. which tries to be reverse
# of what's beeen done. For example clean-generate cleans all the generated files
# And glean-generate-grpc cleans up all the generated grpc files

.PHONY: clean
clean: clean-generate clean-build

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
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/$$SERVICE_NAME ./services/$$SERVICE_PATH

.PHONY: build-deps-go
build-deps-go: $(wildcard services/**/*.go) $(wildcard generated/grpc/go/**/*.go)


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

.PHONY: generate
generate: generate-grpc generate-swagger

.PHONY: generate-swagger
generate-swagger: clean-generate-swagger
	mkdir -p ./generated/swagger/go

	docker pull $(SWAGGER_IMAGE_NAME)

	for spec in `find ./generated/swagger/spec -iname *swagger.json`; do \
		echo "Generating swagger spec for : $$spec" ; \
		docker run --rm -v `pwd`:${PROJECT_PATH} -w ${PROJECT_PATH} $(SWAGGER_IMAGE_NAME) validate $${spec} ; \
		export service_spec_dir=`dirname $${spec}`; \
		export service_name=`basename $${service_spec_dir}`; \
		docker run --rm -v `pwd`:${PROJECT_PATH} -w ${PROJECT_PATH} $(SWAGGER_IMAGE_NAME) generate client -f $${spec} -A service -t ./generated/swagger/go/$${service_name}; \
	done

.PHONY: clean-generate-swagger
clean-generate-swagger:
	rm -rf ./generated/swagger/go

$(GRPC_TARGETS): generated/grpc/go/%: ./contracts/%
	echo "Generating GRPC go packages $@ From $<"
	mkdir -p $@
	mkdir -p ./generated/swagger/spec

	docker run -v `pwd`:${PROJECT_PATH} \
		-w ${PROJECT_PATH} \
		${GRPC_IMAGE_NAME} \
		protoc \
			--go_out=plugins=grpc:/go/src \
			-I ./contracts -I. \
			-I/go/src \
			-I/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
			--grpc-gateway_out=/go/src \
			--swagger_out=logtostderr=true:${PROJECT_PATH}/generated/swagger/spec \
			`find $< -iname *.proto`

	find ./generated/swagger/spec -iname *swagger.json | grep -v service | xargs -I {} rm {}
	find ./generated/swagger -type d -empty | xargs -I {} rm -r {}
	find ./contracts -iname *service.proto | xargs -I {} grep -L "option (google.api.http)" {} | awk -F / '{print $$3 }' |  xargs -I {} rm -rf ./generated/swagger/spec/{}


$(GRPC_PYTHON_TARGETS): pyservices/generated/%: ./contracts/%
	echo "Generating GRPC Python packages $@ From $<"
	mkdir -p $@

	docker run -v `pwd`:${PROJECT_PATH} \
		-w ${PROJECT_PATH} \
		${GRPC_IMAGE_NAME} \
		python3 -m grpc_tools.protoc \
		-I/go/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
		-I/go/src \
		-I ./contracts \
		--python_out=`dirname $@` \
		--grpc_python_out=`dirname $@` \
		`find $< -iname *.proto`

	touch $@/__init__.py


.PHONY: generate-grpc-docker
generate-grpc-docker:
	docker build -t ${GRPC_IMAGE_NAME} -f docker/Dockerfile.grpc .

.PHONY: generate-grpc
generate-grpc: generate-grpc-python generate-grpc-go


.PHONY: generate-grpc-python
generate-grpc-python: clean-generate-grpc-python generate-grpc-docker $(GRPC_PYTHON_TARGETS)
	touch pyservices/generated/__init__.py


.PHONY: generate-grpc-go
generate-grpc-go: clean-generate-grpc-go generate-grpc-docker $(GRPC_TARGETS)


.PHONY: clean-generate-grpc-python
clean-generate-grpc-python:
	rm -rf pyservices/generated

.PHONY: clean-generate-grpc-go
clean-generate-grpc-go:
	rm -rf generated


.PHONY: clean-generate-grpc
clean-generate-grpc: clean-generate-grpc-python clean-generate-grpc-go

.PHONY: clean-generate
clean-generate: clean-generate-grpc clean-generate-swagger