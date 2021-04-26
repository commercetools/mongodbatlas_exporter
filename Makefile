DOCKER_IMAGE_NAME ?= mongodbatlas-exporter
DOCKER_IMAGE_TAG ?= 0.0.1
CGO_ENABLED=0 GOOS=linux GOARCH=amd64

.PHONY: test

test:
	CGO_ENABLED=1 go test -race ./...

docker:
	@echo ">> building docker image"
	docker build -t "$(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)" .

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mongodbatlas_exporter .