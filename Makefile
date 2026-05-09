.PHONY: generate verify-generate

# generate runs go generate to produce generated code from vendored k8s OpenAPI schemas
generate:
	go generate ./...

# verify-generate runs generate and verifies no diff (for CI)
verify-generate: generate
	git diff --exit-code