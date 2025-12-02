BUILD_DATE ?= "$(shell date +"%Y-%m-%dT%H:%M")"
GIT_SHA=$(shell git rev-parse --short=7 HEAD)
REGISTRY ?= registry.cn-hangzhou.aliyuncs.com/dspo
IMAGE_TAG ?= dev

KIND_NAME ?= go-example-e2e
CLUSTER_NAME ?= go-example-e2e
E2E_NAMESPACE ?= go-example-e2e

export KUBECONFIG = /tmp/$(CLUSTER_NAME).kubeconfig

GOOS ?= linux
GOARCH ?= arm64

ifeq ($(shell uname -s),Darwin)
	GOOS = darwin
endif

ifeq ($(shell uname -m),arm64)
	GOARCH = arm64
endif
ifeq ($(shell uname -m), aarch64)
	GOARCH = arm64
endif

gofmt: ## Apply go fmt
	@gofmt -w -r 'interface{} -> any' .
	@gofmt -w -r 'ginkgo.FIt -> ginkgo.It' test
	@gofmt -w -r 'ginkgo.FContext -> ginkgo.Context' test
	@gofmt -w -r 'ginkgo.FDescribe -> ginkgo.Describe' test
	@gofmt -w -r 'ginkgo.FDescribeTable -> ginkgo.DescribeTable' test
	@go fmt ./...
.PHONY: gofmt

lint: gofmt
	go mod tidy
	go vet ./...
	# todo: add more lint like golangci-lint

build-app-image: lint
	docker build -f Dockerfile/app.Dockerfile -t go-example-app:${IMAGE_TAG} .

build-ginkgo-image: lint
	docker build -f Dockerfile/ginkgo.Dockerfile -t ginkgo:dev .

.PHONY: kind-up
kind-up:
	@kind get clusters 2>&1 | grep -v $(KIND_NAME) \
		&& kind create cluster --name $(KIND_NAME) --image docker.cnb.cool/dspo-group/go-example2/node:v1.34.0 \
		|| echo "kind cluster already exists"
	@kind get kubeconfig --name $(KIND_NAME) > $$KUBECONFIG
	kubectl wait --for=condition=Ready nodes --all

.PHONY: kind-down
kind-down:
	@kind get clusters 2>&1 | grep $(KIND_NAME) \
		&& kind delete cluster --name $(KIND_NAME) \
		|| echo "kind cluster does not exist"

.PHONY: kind-load-images
kind-load-images:
	@kind load docker-image mysql:lts --name $(KIND_NAME)
	@kind load docker-image go-example-app:dev --name $(KIND_NAME)
	@kind load docker-image ginkgo:dev --name $(KIND_NAME)

e2e-run:
	@kubectl delete deployment -l testGroup=application --all-namespaces
	@kubectl apply -f test/framework/manifests/namespace.yaml
	@kubectl apply -f test/framework/manifests/configmap.yaml
	@kubectl apply -f test/framework/manifests/ginkgo.yaml
	@kubectl -n $(E2E_NAMESPACE) delete job ginkgo-runtime --ignore-not-found
	@kubectl apply -f test/framework/manifests/ginkgo.job.yaml
	@kubectl -n $(E2E_NAMESPACE) wait --for=condition=complete job/ginkgo-runtime --timeout=30m
	@kubectl -n $(E2E_NAMESPACE) logs job/ginkgo-runtime -f

e2e: kind-up build-app-image build-ginkgo-image kind-load-images e2e-run

e2e-ginkgo: build-ginkgo-image kind-load-images e2e-run
