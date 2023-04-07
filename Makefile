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
	ko build $(POPULATOR_MAIN_PATH) $(CONTROLLER_MAIN_PATH) \
		--platform=linux/amd64,linux/arm64 \
		--base-import-paths

resolve:
	ko resolve --base-import-paths -f ./yaml/deploy.yaml

apply:
	ko apply --base-import-paths -f ./yaml/deploy.yaml

delete:
	ko delete -f ./yaml/deploy.yaml
	ko delete -f ./yaml/rbac.yaml
