FROM golang:1.17.3 as builder

WORKDIR /tmp/gql

COPY . .

ARG BUILDER
ARG VERSION

ENV GQL_BUILDER=${BUILDER}
ENV GQL_VERSION=${VERSION}

RUN apt-get install make git gcc && \
    make build_deps && \
    make

FROM alpine:latest

WORKDIR /app

COPY --from=builder /tmp/gql/bin/gql .

ENTRYPOINT ["./gql"]
