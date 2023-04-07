REMOTE_REGISTRY ?= isim

CONTROLLER_MAIN_PATH ?= ./cmd/cbt-controller
POPULATOR_MAIN_PATH := ./cmd/cbt-populator

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: codegen
codegen:
	./hack/update-codegen.sh

build:
	ko build $(POPULATOR_MAIN_PATH) $(CONTROLLER_MAIN_PATH) \
		--local \
		--platform=linux/amd64,linux/arm64 \
		--base-import-paths

push:
	KO_DOCKER_REPO=$(REMOTE_REGISTRY) \
	ko build $(POPULATOR_MAIN_PATH) $(CONTROLLER_MAIN_PATH) \
		--platform=linux/amd64,linux/arm64 \
		--base-import-paths

resolve:
	KO_DOCKER_REPO=$(REMOTE_REGISTRY) \
	ko resolve --base-import-paths -f ./yaml/deploy.yaml

apply:
	KO_DOCKER_REPO=$(REMOTE_REGISTRY) \
	ko apply --base-import-paths -f ./yaml/deploy.yaml

delete:
	ko delete -f ./yaml/deploy.yaml
