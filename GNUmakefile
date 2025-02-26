default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./internal/provider

testacc:
	TF_ACC=1 go test -v -cover ./internal/provider

.PHONY: fmt lint test testacc build install generate
