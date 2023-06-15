#!/bin/bash

mkdir -p out

go build -o out/server cli/server.go
go build -o out/test test/main.go

# go build -o out/client cli/client.go
