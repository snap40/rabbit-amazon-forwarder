FROM golang:1.16.3-alpine3.13 AS builder

RUN apk add --no-cache curl git openssh \
 && adduser -D -g '' appuser

COPY . /go/src/github.com/snap40/rmq-aws-forwarder
WORKDIR /go/src/github.com/snap40/rmq-aws-forwarder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN  go mod tidy \
     && go mod verify \
     && go mod vendor

RUN go build -ldflags="-w -s" -o /go/src/github.com/snap40/rmq-aws-forwarder/rabbit-amazon-forwarder

FROM alpine:3.13

RUN adduser -D -g '' appuser

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/snap40/rmq-aws-forwarder/rabbit-amazon-forwarder /app/forwarder
USER appuser

ENTRYPOINT ["/app/forwarder"]
