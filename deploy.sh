#!/bin/sh

CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o build/main main.go
#scp build/main root@167.99.107.244:/root
