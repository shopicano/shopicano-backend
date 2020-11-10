FROM golang:alpine AS builder

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN apk add git openssh

ENV GOPATH=/go

ENV GOOS="linux"
ENV GOARCH="amd64"
ENV GO111MODULE=on

COPY . $GOPATH/src/github.com/shopicano/shopicano-backend
WORKDIR $GOPATH/src/github.com/shopicano/shopicano-backend

RUN go get .
RUN go build -v -o yallawebsites
RUN mv yallawebsites /go/bin/yallawebsites

FROM alpine

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

WORKDIR /root

COPY --from=builder /go/bin/yallawebsites /usr/local/bin/yallawebsites

ENTRYPOINT ["yallawebsites"]
