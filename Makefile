BINARY_NAME=apm

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
	go build -o ${BINARY_NAME} cmd/apm/*

clean:
	go clean
	rm ${BINARY_NAME}