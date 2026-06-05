#!/usr/bin/env bash
go build -o ../cmd/echo/echo ../cmd/echo/echo.go
../cmd/echo/echo "the quick brown fox" > fox.txt
../cmd/echo/echo "jumps over the lazy dog" > dog.txt

go build -o ../cmd/cat/cat ../cmd/cat/cat.go
../cmd/cat/cat fox.txt dog.txt