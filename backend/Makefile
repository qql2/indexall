GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

.PHONY: init
# init env
init:
	go install github.com/google/wire/cmd/wire@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest

.PHONY: api
# generate api proto code
api:
	protoc --proto_path=./api \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./api \
	       --go-http_out=paths=source_relative:./api \
	       --go-grpc_out=paths=source_relative:./api \
	       $(shell find api -name '*.proto')

.PHONY: config
# generate config proto code
config:
	protoc --proto_path=./internal/conf \
	       --proto_path=./third_party \
	       --go_out=paths=source_relative:./internal/conf \
	       internal/conf/conf.proto

.PHONY: wire
# generate wire
wire:
	cd cmd/server && wire

.PHONY: build
# build
build:
	mkdir -p bin/ && CGO_ENABLED=1 go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./cmd/server

.PHONY: generate
# generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
# all
all:
	make api
	make config
	make generate
	make build

.PHONY: run
# run application
run:
	go run ./cmd/server -conf=./configs

.PHONY: test
# run tests
test:
	go test -v ./...

.DEFAULT_GOAL := all
