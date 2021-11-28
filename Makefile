BINARY_NAME=apm

all: lint test build

lint:
	golint ./... > report.log

test:
	go test -cover -v ./internal/**
	# go test -v ./cmd/**

build:
	go build -o ${BINARY_NAME} cmd/apm/*

clean:
	go clean
	rm ${BINARY_NAME}