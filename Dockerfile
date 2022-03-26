FROM golang:1.18-alpine as builder

WORKDIR /tmp/gql

COPY . .

ARG BUILDER
ARG VERSION

ENV GQL_BUILDER=${BUILDER}
ENV GQL_VERSION=${VERSION}

RUN apk add --no-cache make git gcc && \
    make build_deps && \
    make linux

FROM ubuntu:latest

WORKDIR /app

COPY --from=builder /tmp/gql/bin/gql .

ENTRYPOINT ["./gql"]
