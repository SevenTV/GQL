all: linux

BUILDER := "unknown"
VERSION := "unknown"

ifeq ($(origin GQL_BUILDER),undefined)
	BUILDER = $(shell git config --get user.name);
else
	BUILDER = ${GQL_BUILDER};
endif

ifeq ($(origin GQL_VERSION),undefined)
	VERSION = $(shell git rev-parse HEAD);
else
	VERSION = ${GQL_VERSION};
endif

linux:
	packr2
	GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.Unix=$(shell date +%s)' -X 'main.User=${BUILDER}'" -o bin/gql .
	packr2 clean

lint:
	staticcheck ./...
	go vet ./...
	golangci-lint run

deps:
	go mod download
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/gobuffalo/packr/v2/packr2@latest

test:
	go test -count=1 -cover ./...
