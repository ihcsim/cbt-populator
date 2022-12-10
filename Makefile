.PHONY: test
test:
	go test ./...

.PHONY: codegen
codegen:
	./hack/update-codegen.sh
