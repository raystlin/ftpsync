#!/bin/bash

OS=linux
ARCH=arm

docker run --rm -v /home/user/go:/go -e GOOS=$OS -e GOARCH=$ARCH golang:1.13.1-alpine3.10 go build -o /go/src/github.com/raystlin/ftpsync/dist/ftpsync github.com/raystlin/ftpsync
