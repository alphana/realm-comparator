FROM golang:1.19-alpine3.16.2 as builder

ENV GO111MODULE=on

WORKDIR /go/src/github.com/soss-sig/keycloak-fapi
COPY *.go ./

RUN go mod init
RUN go build -o realm-comparator-server *.go

FROM alpine:3.16.2

# Install consent-server for testing
COPY --from=builder /go/src/github.com/soss-sig/keycloak-fapi/consent-server /usr/local/sbin/

ENTRYPOINT [ "realm-comparator-server" ]