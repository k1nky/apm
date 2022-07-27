BINARY_NAME=apm
VERSION=$(shell git describe --tags)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

GOLDFLAGS = -s
GOLDFLAGS += -X main.BuildVersion=${VERSION}
GOLDFLAGS += -X main.BuildTarget=${GOOS}/${GOARCH}

all: lint test build

lint:
	golint ./... | tee report.log

cover:
	go test -v -cover ./internal/** -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

test:
	go test -cover -v ./internal/**
	go test -v ./cmd/**

build:
	go build -ldflags="${GOLDFLAGS}" -o ${BINARY_NAME} cmd/apm/*

clean:
	go clean
	rm ${BINARY_NAME}