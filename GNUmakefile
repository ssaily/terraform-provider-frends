default: build

build:
	go build -v ./...

install: build
	go install -v ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -cover -timeout=120m ./...

generate:
	cd tools; go generate ./...

.PHONY: build install fmt test testacc generate
