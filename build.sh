#!/bin/bash

mkdir -p dist/

go build -o dist/ezfaas main.go

ls -lh dist/
