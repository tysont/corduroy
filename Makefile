SHELL = /bin/bash
PROJECT_NAME = corduroy
PROJECT_PATH = github.com/tysont/${PROJECT_NAME}
FULL_PROJECT_PATH = ${GOPATH}/src/${PROJECT_PATH}
BUILD_DIRECTORY = bin
BUILD_PATH = ${BUILD_DIRECTORY}/${PROJECT_NAME}
CONTAINER_TAG = ${PROJECT_NAME}/build:latest
CONTAINER_PROJECT_PATH = /opt/go/src/${PROJECT_PATH}

.PHONY: all clean test build run prepare-container test-container build-container run-container
all: clean test build
all-container: prepare-container test-container build-container run-container

clean:
	rm -rf bin

build:
	GOOS=linux GOARCH=amd64 go build -o ${BUILD_PATH} ${PROJECT_PATH}/cmd

test:
	pwd
	ls vendor/github.com/
	go test ./...

run:
	${BUILD_PATH}

prepare-container:
	docker build --rm -t ${CONTAINER_TAG} .

test-container:
	docker run -it --rm -v ${FULL_PROJECT_PATH}:${CONTAINER_PROJECT_PATH} --workdir ${CONTAINER_PROJECT_PATH} ${CONTAINER_TAG} make test

build-container:
	docker run -it --rm -v ${FULL_PROJECT_PATH}:${CONTAINER_PROJECT_PATH} --workdir ${CONTAINER_PROJECT_PATH} ${CONTAINER_TAG} make build

run-container:
	docker run -it --rm -v ${FULL_PROJECT_PATH}:${CONTAINER_PROJECT_PATH} -p 8080:8080 --workdir ${CONTAINER_PROJECT_PATH} ${CONTAINER_TAG} make run