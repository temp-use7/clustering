SHELL := /bin/sh

.PHONY: all build clusterd nodeagent clustectl tidy proto

all: build

build: clusterd nodeagent clustectl

clusterd:
	go build -o bin/clusterd ./cmd/clusterd

nodeagent:
	go build -o bin/nodeagent ./cmd/nodeagent

clustectl:
	go build -o bin/clustectl ./cmd/clustectl

tidy:
	go mod tidy

proto:
	@echo "(Optional) generate gRPC stubs with buf/protoc; placeholder target"



