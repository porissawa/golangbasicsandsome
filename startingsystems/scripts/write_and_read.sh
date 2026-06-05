#!/usr/bin/env bash
go build -o ../bin/echo ../cmd/echo/echo.go
../bin/echo "the quick brown fox" > fox.txt
../bin/echo "jumps over the lazy dog" > dog.txt

go build -o ../bin/cat ../cmd/cat/cat.go
../bin/cat fox.txt dog.txt