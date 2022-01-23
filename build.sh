#!/bin/bash

mkdir -p dist/

go build -o dist/ezfaas internal/main.go

ls -lh dist/
