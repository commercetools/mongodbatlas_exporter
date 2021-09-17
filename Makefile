DOCKER_IMAGE_NAME ?= mongodbatlas-exporter
CGO_ENABLED=0 GOOS=linux GOARCH=amd64
BIN_DIR=.
BIN=mongodbatlas_exporter
VERSION=$(shell git describe --exact-match --tags HEAD || echo "latest")
REVISION=$(shell git rev-parse HEAD)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)

.PHONY: test

test:
	CGO_ENABLED=1 go test -race ./...

docker:
	@echo ">> building docker image"
	docker build -t "$(DOCKER_IMAGE_NAME):$(VERSION)" .

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo \
	-ldflags="-w -s -X github.com/prometheus/common/version.Branch=${BRANCH} -X github.com/prometheus/common/version.Revision=${REVISION} -X github.com/prometheus/common/version.Version=${VERSION}" \
	-o $(BIN_DIR)/$(BIN)
