#!/bin/bash

set -e

echo ">==( go mod verify )==="
go mod verify
echo "<==(      OK       )==="
echo "" 

echo ">==( gofmt -d . )==="
fmt=$(gofmt -l .)
if [ "$fmt" != "" ]; then
    echo $fmt
    exit 1
fi
echo "<==(     OK     )==="
echo ""

echo ">==( go vet ./... )==="
go vet ./...
echo "<==(      OK      )==="
echo ""

echo ">==( go test --cover --race -timeout 30s ./... )==="
go test --cover --race -timeout 30s ./...
echo "<==(      OK       )==="
echo ""
