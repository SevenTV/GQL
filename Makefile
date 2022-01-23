all: gql linux

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
	yarn prettier --write .

deps: go_installs
	go mod download
	yarn

build_deps:
	go install github.com/gobuffalo/packr/v2/packr2@latest
	go install github.com/99designs/gqlgen@latest
	go install github.com/vektah/dataloaden@latest

go_installs: build_deps
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

gql:
	gqlgen

	cd graph/loaders && dataloaden UserLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/model.User"
	cd graph/loaders && dataloaden BatchUserLoader string "[]*github.com/SevenTV/GQL/graph/model.User"

	cd graph/loaders && dataloaden EmoteLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/model.Emote"
	cd graph/loaders && dataloaden BatchEmoteLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "[]*github.com/SevenTV/GQL/graph/model.Emote"

	cd graph/loaders && dataloaden EmoteSetLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/model.EmoteSet"

	cd graph/loaders && dataloaden RoleLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/model.Role"

	cd graph/loaders && dataloaden ConnectionLoader string "*github.com/SevenTV/GQL/graph/model.UserConnection"

	cd graph/loaders && dataloaden ReportLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/model.Report"
	cd graph/loaders && dataloaden BatchReportLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "[]*github.com/SevenTV/GQL/graph/model.Report"

test:
	go test -count=1 -cover ./...
