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
	GOOS=linux GOARCH=amd64 go build -v -ldflags "-X 'main.Version=${VERSION}' -X 'main.Unix=$(shell date +%s)' -X 'main.User=${BUILDER}'" -o bin/gql .

lint:
# Disable staticcheck because they do not yet support generics.
# Remove the comment once support is added
#	staticcheck ./...
	go vet ./...
	golangci-lint run
	yarn prettier --write .

deps: go_installs
	go mod download
	yarn

build_deps:
	go install github.com/99designs/gqlgen@v0.17.2
	go install github.com/seventv/dataloaden@cc5ac4900

go_installs: build_deps
	go install honnef.co/go/tools/cmd/staticcheck@master
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

gql:
	gqlgen --config ./gqlgen.v3.yml

	cd graph/v3/loaders && dataloaden UserLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/v3/model.User"
	cd graph/v3/loaders && dataloaden BatchUserLoader string "[]*github.com/SevenTV/GQL/graph/v3/model.User"

	cd graph/v3/loaders && dataloaden EmoteLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/v3/model.Emote"
	cd graph/v3/loaders && dataloaden BatchEmoteLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "[]*github.com/SevenTV/GQL/graph/v3/model.Emote"

	cd graph/v3/loaders && dataloaden EmoteSetLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/v3/model.EmoteSet"
	cd graph/v3/loaders && dataloaden BatchEmoteSetLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "[]*github.com/SevenTV/GQL/graph/v3/model.EmoteSet"

	cd graph/v3/loaders && dataloaden RoleLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/v3/model.Role"

	cd graph/v3/loaders && dataloaden ConnectionLoader string "*github.com/SevenTV/GQL/graph/v3/model.UserConnection"

	cd graph/v3/loaders && dataloaden ReportLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "*github.com/SevenTV/GQL/graph/v3/model.Report"
	cd graph/v3/loaders && dataloaden BatchReportLoader "go.mongodb.org/mongo-driver/bson/primitive.ObjectID" "[]*github.com/SevenTV/GQL/graph/v3/model.Report"

	gqlgen --config ./gqlgen.v2.yml

	cd graph/v2/loaders && dataloaden UserLoader "string" "*github.com/SevenTV/GQL/graph/v2/model.User"
	cd graph/v2/loaders && dataloaden UserEmotesLoader "string" "[]*github.com/SevenTV/GQL/graph/v2/model.Emote"

	cd graph/v2/loaders && dataloaden EmoteLoader "string" "*github.com/SevenTV/GQL/graph/v2/model.Emote"

test:
	go test -count=1 -cover ./...
