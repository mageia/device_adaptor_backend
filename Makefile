.PHONY: all bin build assets build-alpine frontend clean test help default

BIN_NAME=device_adaptor
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME := "harbor.leaniot.cn/mos/data/device_adaptor"

default: all

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

all: bin

run:
	cd cmd && ./${BIN_NAME}

pi:
	cd cmd && CGO_ENABLED=1 CC=/Volumes/arm-linux/bin/arm-none-linux-gnueabi-gcc GOARM=6 GOARCH=arm GOOS=linux go build -o ../deviceAdaptorLinux

bin:
	@echo "building exec ${BIN_NAME}"
	cd cmd && go build -o ${BIN_NAME} .

assets:
	@echo "building assets"
	rm -rf assets/dist
	cp -r frontend/dist assets
	statik -src=assets

frontend:
	@echo "building frontend"
	rm -rf frontend/dist
	cd frontend && npm install --registry=https://registry.npm.taobao.org --verbose && npm run build

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
	@test ! -e cmd/${BIN_NAME} || rm cmd/${BIN_NAME}

test:
	go test ./...
