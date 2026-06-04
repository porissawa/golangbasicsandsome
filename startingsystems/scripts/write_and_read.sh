#!/usr/bin/env bash
../cmd/echo/echo "the quick brown fox" > fox.txt
../cmd/echo/echo "jumps over the lazy dog" > dog.txt
../cmd/cat/cat fox.txt dog.txt