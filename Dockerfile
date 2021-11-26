FROM golang:1.17.3-alpine as builder

WORKDIR /tmp/gql

COPY . .

ARG BUILDER
ARG VERSION

ENV GQL_BUILDER=${BUILDER}
ENV GQL_VERSION=${VERSION}

RUN apk add --no-cache make git && \
    make linux

FROM alpine:latest

WORKDIR /app

COPY --from=builder /tmp/gql/bin/gql .

ENTRYPOINT ["./gql"]
