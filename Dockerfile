FROM --platform=amd64 golang:1.16.3-alpine3.13 AS builder

RUN apk add --no-cache curl git openssh \
 && adduser -D -g '' appuser

COPY . /go/src/github.com/AirHelp/rabbit-amazon-forwarder
WORKDIR /go/src/github.com/AirHelp/rabbit-amazon-forwarder

ENV GO111MODULE=on
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0

RUN  go mod tidy \
     && go mod verify \
     && go mod vendor

RUN go build -ldflags="-w -s" -o /go/src/github.com/AirHelp/rabbit-amazon-forwarder/rabbit-amazon-forwarder

FROM --platform=amd64 alpine:3.13

RUN adduser -D -g '' appuser

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/AirHelp/rabbit-amazon-forwarder/rabbit-amazon-forwarder /app/forwarder
ADD https://www.amazontrust.com/repository/SFSRootCAG2.pem /app/SFSRootCAG2.pem
RUN chmod a+r /app/SFSRootCAG2.pem
ENV CA_CERT_FILE=/app/SFSRootCAG2.pem

USER appuser

ENTRYPOINT ["/app/forwarder"]
