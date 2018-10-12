.PHONY: build build-alpine clean test help default

BIN_NAME=deviceAdaptor
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME := "harbor.leaniot.cn/mos/device_adaptor"

default: help

help:
	@echo 'Management commands for device_adaptor:'
	@echo
	@echo 'Usage:'
	@echo '    make package         Build final docker image with just the go binary inside'
	@echo '    make tag             Tag image created by package with latest, git commit and version'
	@echo '    make test            Run tests on a compiled project.'
	@echo '    make push            Push tagged images to registry'
	@echo '    make clean           Clean the directory tree.'
	@echo

package:
	@echo "building image ${BIN_NAME} $(GIT_COMMIT)"
	docker build -t $(IMAGE_NAME):local .

tag:
	@echo "Tagging: latest $(GIT_COMMIT)"
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):$(GIT_COMMIT)
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):latest

push: tag
	@echo "Pushing docker image to registry: latest $(GIT_COMMIT)"
	docker push $(IMAGE_NAME):$(GIT_COMMIT)
	docker push $(IMAGE_NAME):latest

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}

test:
	go test ./...
