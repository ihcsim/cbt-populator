LOCAL_REGISTRY = ko.local
REMOTE_REGISTRY ?= isim

CONTROLLER_MAIN_PATH ?= ./cmd/cbt-controller
POPULATOR_MAIN_PATH := ./cmd/cbt-populator

build-%:
	if [ "$*" = "remote" ]; then \
    export KO_DOCKER_REPO=$(REMOTE_REGISTRY) ;\
	else  \
    export KO_DOCKER_REPO=$(LOCAL_REGISTRY) ;\
	fi && \
	ko build $(POPULATOR_MAIN_PATH) $(CONTROLLER_MAIN_PATH) \
		--platform=linux/amd64,linux/arm64 \
		--base-import-paths

build: build-local

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: codegen
codegen:
	./hack/update-codegen.sh
