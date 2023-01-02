LOCAL_REGISTRY = ko.local
REMOTE_REGISTRY ?= isim

GO_MAIN_PATH ?= ./cmd/populator

build-%:
	if [ "$*" = "remote" ]; then \
    export KO_DOCKER_REPO=$(REMOTE_REGISTRY) ;\
	else  \
    export KO_DOCKER_REPO=$(LOCAL_REGISTRY) ;\
	fi && \
	ko build $(GO_MAIN_PATH) \
		--base-import-paths
build: build-local

.PHONY: test
test:
	go test -cover -race ./...

.PHONY: codegen
codegen:
	./hack/update-codegen.sh
