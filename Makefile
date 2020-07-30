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
COMPOSE=docker-compose -p $(CI_JOB_ID)
COMMON_SH=source ./scripts/common.sh &&
GINKGO_FOCUS?=integration

HELM_IMAGE_NAME=$(REPOSITORY_NAME)/eskit-helm:latest

KUBECONFIG_DIR=/tmp/kube
KUBECONFIG_PATH=/tmp/kube/config

COMPOSE_FILE=docker-compose-unit.yml

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
DEPLOY_TARGETS += eventstore
DEPLOY_TARGETS += consumers/metrics
DEPLOY_TARGETS += crudstore
DEPLOY_TARGETS += consumerstore
DEPLOY_TARGETS += users

K8S_SERVICES =
K8S_SERVICES += eventstore
K8S_SERVICES += crudstore
K8S_SERVICES += consumerstore
K8S_SERVICES += users


GO_BUILD_SERVICES =
GO_BUILD_SERVICES += eventstore/cmd/eventstore
GO_BUILD_SERVICES += eventstore/cmd/eventstoregw
GO_BUILD_SERVICES += crudstore/cmd/crudstore
GO_BUILD_SERVICES += crudstore/cmd/crudstoregw
GO_BUILD_SERVICES += consumerstore/cmd/consumerstore
GO_BUILD_SERVICES += consumerstore/cmd/consumerstoregw
GO_BUILD_SERVICES += users/cmd/users
GO_BUILD_SERVICES += users/cmd/usersgw
GO_BUILD_SERVICES += consumers/metrics
GO_BUILD_TARGETS = $(addprefix ./.bin/, $(GO_BUILD_SERVICES))


.PHONY: all
all: build

.PHONY: build
build: build-compose-go build-docker

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
	go mod vendor


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
build-deps-go: $(wildcard vendor/**/*.go) $(wildcard services/**/*.go) $(wildcard generated/grpc/go/**/*.go)


.PHONY: build-docker
build-docker: build-docker-helm
	@eval $(shell source ./scripts/helm.sh && print_next_version $(HELM_IMAGE_NAME) `pwd` charts/eventstore); \
	env |grep APP_VERSION; \
	for service in $(DEPLOY_TARGETS); do \
		IMAGE_NAME=`basename $$service`; \
		docker build --tag $(REPOSITORY_NAME)/eskit-$$IMAGE_NAME:$$APP_VERSION -f ./services/$$service/Dockerfile . ; \
	done

.PHONY: build-docker-helm
build-docker-helm:
	docker build -t $(HELM_IMAGE_NAME) -f docker/Dockerfile.helm .

.PHONY: build-minikube-docker
build-minikube-docker:
	@eval $$(minikube docker-env) ;\
	$(MAKE) build-docker-helm ;\
	$(MAKE) build-docker

.PHONY: clean-build-go
clean-build-go:
	for build_target in $(GO_BUILD_SERVICES); do \
		rm -rf ./bin ; \
	done

.PHONY: clean-build
clean-build: clean-build-go


# Deployment related targets

.PHONY: deploy
deploy: deploy-minikube

.PHONY: deploy-minikube
deploy-minikube: build-compose-go build-minikube-docker deploy-minikube-helm

.PHONY: deploy-minikube-helm
deploy-minikube-helm: deploy-minikube-setup deploy-helm

.PHONY: deploy-minikube-setup
deploy-minikube-setup:
	$(KUBECTL_SH) config use-context minikube
	$(MAKE) deploy-helm-deps

.PHONY: deploy-helm-deps
deploy-helm-deps:
	$(HELM_SH) init --client-only
	$(HELM_SH) ls --all --short | xargs -n1 $(HELM_SH) delete --purge
	$(HELM_SH) upgrade --wait --install --recreate-pods --debug eskit-db stable/postgresql --namespace=${NAMESPACE} --set postgresqlDatabase=eventsourcing --set persistence.enabled=false
	$(HELM_SH) upgrade --install --recreate-pods --debug eskit-prom stable/prometheus --namespace=${NAMESPACE}  -f charts/prometheus/values.yaml
	$(HELM_SH) upgrade --wait --install --recreate-pods --debug eskit-grafanadash charts/grafanadash --namespace=${NAMESPACE}
	$(HELM_SH) upgrade --install --recreate-pods --debug eskit-grafana stable/grafana --namespace=${NAMESPACE}  -f charts/grafana/values.yaml

.PHONY: deploy-do
deploy-do: deploy-do-helm

.PHONY: deploy-do-helm
deploy-do-helm: deploy-do-setup
	test -n "$${KUBECONFIG_PAYLOAD}" || (echo "KUBECONFIG_PAYLOAD is required" ; exit 1)
	mkdir -p $(KUBECONFIG_DIR)
	echo $${KUBECONFIG_PAYLOAD} | $(BASE64_DECODE) > $(KUBECONFIG_PATH)
	$(MAKE) deploy-cloud-helm; \
	rm $(KUBECONFIG_PATH)

.PHONY: deploy-do-setup
deploy-do-setup:
	test -n "$${KUBECONFIG_PAYLOAD}" || (echo "KUBECONFIG_PAYLOAD is required" ; exit 1)
	mkdir -p $(KUBECONFIG_DIR)
	echo $${KUBECONFIG_PAYLOAD} | $(BASE64_DECODE) > $(KUBECONFIG_PATH)
	$(HELM_SH) init --client-only; \
	$(HELM_SH) ls --all --short | grep -v ingress | xargs -n1 $(HELM_SH) delete --purge; \
	$(HELM_SH) upgrade --wait --install --recreate-pods --debug eskit-db stable/postgresql --namespace=${NAMESPACE} --set postgresqlDatabase=eventsourcing --set persistence.enabled=false; \
	$(HELM_SH) upgrade --install --recreate-pods --debug eskit-prom stable/prometheus --namespace=${NAMESPACE}  -f charts/prometheus/values.yaml --set server.ingress.hosts[0]=proms.$(HOSTNAME_SUFFIX); \
	$(HELM_SH) upgrade --wait --install --recreate-pods --debug eskit-grafanadash charts/grafanadash --namespace=${NAMESPACE}; \
	$(HELM_SH) upgrade --install --recreate-pods --debug eskit-grafana stable/grafana --namespace=${NAMESPACE}  -f charts/grafana/values.yaml --set ingress.hosts[0]=grafana.$(HOSTNAME_SUFFIX); \
	$(HELM_SH) upgrade --install --debug nginx-ingress stable/nginx-ingress --set rbac.create=true; \
	rm $(KUBECONFIG_PATH)

.PHONY: deploy-helm-tiller-init
deploy-helm-tiller-init:
	test -n "$${KUBECONFIG_PAYLOAD}" || (echo "KUBECONFIG_PAYLOAD is required" ; exit 1)
	mkdir -p $(KUBECONFIG_DIR)
	echo $${KUBECONFIG_PAYLOAD} | $(BASE64_DECODE) > $(KUBECONFIG_PATH)
	$(KUBECTL_SH) create serviceaccount --namespace kube-system tiller; \
	$(KUBECTL_SH) create clusterrolebinding tiller-cluster-rule --clusterrole=cluster-admin --serviceaccount=kube-system:tiller; \
	$(KUBECTL_SH) patch deploy --namespace kube-system tiller-deploy -p '{"spec":{"template":{"spec":{"serviceAccount":"tiller"}}}}'; \
	$(HELM_SH) init --service-account tiller --upgrade --wait; \
	rm $(KUBECONFIG_PATH)


.PHONY: deploy-helm
deploy-helm:
	helm init --client-only

	for dir in $(DEPLOY_TARGETS); do \
		IMAGE_NAME=`basename $$dir`; \
		source ./scripts/helm.sh && set_chart_version $(HELM_IMAGE_NAME) `pwd` charts/$$IMAGE_NAME; \
		RELEASE_NAME="eskit-$$IMAGE_NAME"; \
		$(HELM_SH) upgrade --install --recreate-pods --debug $$RELEASE_NAME ./charts/$$IMAGE_NAME --namespace=${NAMESPACE} --set image.pullPolicy=Never --set service.type=NodePort; \
	done

.PHONY: deploy-cloud-helm
deploy-cloud-helm:
	$(HELM_SH) init --client-only
	$(KUBECTL_SH) config view
	for dir in $(DEPLOY_TARGETS); do \
		IMAGE_NAME=`basename $$dir`; \
		source ./scripts/helm.sh && set_chart_version $(HELM_IMAGE_NAME) `pwd` charts/$$IMAGE_NAME; \
		RELEASE_NAME="eskit-$$IMAGE_NAME"; \
		$(HELM_SH) upgrade --install --recreate-pods --debug $$RELEASE_NAME ./charts/$$IMAGE_NAME --namespace=${NAMESPACE} --set ingress.hosts[0]=$$IMAGE_NAME.$(HOSTNAME_SUFFIX) --set image.pullPolicy=Always; \
	done


.PHONY: deploy-helm-set-version
deploy-helm-set-version:
	for dir in $(DEPLOY_TARGETS); do \
		IMAGE_NAME=`basename $$dir`; \
		RELEASE_NAME="eskit-$$IMAGE_NAME"; \
		source ./scripts/helm.sh && set_chart_version $(HELM_IMAGE_NAME) `pwd` charts/$$IMAGE_NAME; \
	done

# TODO change this will be different !!!
.PHONY: deploy-gcloud
deploy-gcloud: deploy-gcloud-setup deploy-cloud-helm

.PHONY: deploy-gcloud-setup
deploy-gcloud-setup:
	$(KUBECTL_SH) config use-context $(GCLOUD_CONTEXT)

.PHONY: deploy-compose
deploy-compose: clean-build build-compose-go
	$(COMMON_SH) source ./scripts/compose.sh && compose_deploy $(CI_JOB_ID)

# TEST targets will be here

.PHONY: test
# test: build test-compose-go test-compose-py
test: build test-compose-go

.PHONY: test-compose-go
test-compose-go:
	export COMPOSE_FILE=$(COMPOSE_FILE); \
	$(COMMON_SH) source ./scripts/compose.sh && compose_tests_golang $(CI_JOB_ID)

.PHONY: test-compose-py
test-compose-py:
	$(COMMON_SH) source ./scripts/compose.sh && compose_tests_pytest $(CI_JOB_ID)

.PHONY: test-minikube
test-minikube: deploy-minikube test-minikube-wait-svc
	@eval $(shell $(MAKE) test-minikube-env ); \
	env ; \
	$(MAKE) test-go-integration

.PHONY: test-minikube-env
test-minikube-env:
	@source ./scripts/minikube.sh && minikube_service_endpoints $(K8S_SERVICES)


.PHONY: test-minikube-wait-svc
test-minikube-wait-svc:
	for service in $(K8S_SERVICES); do \
		source ./scripts/minikube.sh && wait_for_service $$service ; \
	done

.PHONY: test-go
test-go: test-go-unit

.PHONY: test-go-unit
test-go-unit:
	cd services && go test -count=1 -v ./...

.PHONY: test-go-integration
test-go-integration:
	cd tests && ginkgo -r -v --regexScansFilePath=true --focus=$(GINKGO_FOCUS)

.PHONY: test-py
test-py: test-py-unit test-py-integration

.PHONY: test-py-unit
test-py-unit:
	PYTHONPATH=$$PYTHONPATH:`pwd`/pyservices:`pwd`/pyservices/generated:`pwd`/pyservices/store py.test -s --ignore=`pwd`/pyservices/store/tests/integration pyservices

.PHONY: test-py-integration
test-py-integration:
	PYTHONPATH=$$PYTHONPATH:`pwd`/pyservices:`pwd`/pyservices/generated:`pwd`/pyservices/store py.test -s `pwd`/pyservices/store/tests/integration


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

# publish Targets

.PHONY: publish
publish: build publish-dockerhub-docker

.PHONY: publish-docker
publish-docker:
	@eval $(shell source ./scripts/helm.sh && print_next_version $(HELM_IMAGE_NAME) `pwd` charts/eventstore); \
	env |grep APP_VERSION; \
	for dir in $(DEPLOY_TARGETS); do \
		IMAGE_NAME=`basename $$dir`; \
		docker push $(REPOSITORY_NAME)/eskit-$$IMAGE_NAME:$$APP_VERSION; \
	done

.PHONY: publish-dockerhub-docker
publish-dockerhub-docker: publish-dockerhub-docker-env publish-docker


.PHONY: publish-dockerhub-docker-env
publish-dockerhub-docker-env:
	test -n "$${DOCKER_USERNAME}" || (echo "DOCKER_USERNAME is required" ; exit 1)
	test -n "$${DOCKER_PASSWORD}" || (echo "DOCKER_PASSWORD is required" ; exit 1)
	echo $$DOCKER_PASSWORD | docker login --username $$DOCKER_USERNAME --password-stdin

.PHONY: publish-gcloud-docker
publish-gcloud-docker: publish-gcloud-docker-env publish-docker

.PHONY: publish-gcloud-docker-env
publish-gcloud-docker-env:
ifndef GCLOUD_SERVICE_KEY
	$(error GCLOUD_SERVICE_KEY is undefined)
endif
	docker login -u _json_key -p "$$(echo "$$GCLOUD_SERVICE_KEY" | $(BASE64_DECODE))" https://eu.gcr.io


# Some Misc tasks in format of run-{env}-command

.PHONY: run-pyshell
run-pyshell:
	source `pwd`/pyservices/venv/bin/activate && PYTHONPATH=$$PYTHONPATH:`pwd`/pyservices:`pwd`/pyservices/generated:`pwd`/pyservices/store ipython

.PHONY: run-pyvenv
run-pyvenv:
	rm -rf `pwd`/pyservices/venv
	virtualenv `pwd`/pyservices/venv
	source `pwd`/pyservices/venv/bin/activate && pip install -r `pwd`/pyservices/store/requirements.txt

.PHONY: run-version
run-version: build-docker-helm
	@source ./scripts/helm.sh && print_next_version $(HELM_IMAGE_NAME) `pwd` charts/eventstore

.PHONY:run-get-grafana-pass
run-get-grafana-pass:
	$(KUBECTL_SH) get secrets eskit-grafana -o=jsonpath='{.data.admin-password}' | $(BASE64_DECODE) | xargs

.PHONY: run-helm
run-helm:
	test -n "$${KUBECONFIG_PAYLOAD}" || (echo "KUBECONFIG_PAYLOAD is required" ; exit 1)
	mkdir -p $(KUBECONFIG_DIR)
	echo $${KUBECONFIG_PAYLOAD} | $(BASE64_DECODE) > $(KUBECONFIG_PATH)
	export KUBECONFIG=$(KUBECONFIG_PATH); \
	echo $$KUBECONFIG; \
	$(HELM_SH) ls --all --short | grep -v ingress | xargs -n1; \
	rm $(KUBECONFIG_PATH)