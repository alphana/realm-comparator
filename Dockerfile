FROM golang:1.19.0-alpine3.16 as builder

ENV GO111MODULE=on

WORKDIR /go/src/github.com/alphana/
COPY *.go ./

RUN go mod init
RUN go build -o realm-comparator-server *.go

FROM alpine:3.16.2

# Install consent-server for testing

COPY --from=builder /go/src/github.com/alphana/realm-comparator-server /usr/local/sbin/

ENTRYPOINT [ "realm-comparator-server" ]