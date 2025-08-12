# SPECIAL: Hardcoded provider prefix required in this file
default: fmt lint install generate

build:
	go build -v -tags=repoAWX ./...

install: build
	go install -v -tags=repoAWX ./...

lint:
	golangci-lint run

generate:
	go run generate-examples/main.go
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
# SPECIAL: add -tags=repoAWX or -tags=repoAAP
	go test -v -cover -timeout=120s -parallel=10 -tags=repoAWX ./internal/provider

testacc:
# SPECIAL: add -tags=repoAWX or -tags=repoAAP
	TF_ACC=1 go test -tags=repoAWX -v -cover ./internal/provider

.PHONY: fmt lint test testacc build install generate
