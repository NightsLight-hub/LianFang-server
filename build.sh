#!/usr/bin/env bash
currentDir=$(dirname "$0")
programName="lianfang"
if [ "$1" == "arm64" ]; then
  export GOARCH=arm64
else
  export GOARCH=amd64
fi

export CGO_ENABLED=0
export GOOS=linux
echo "Building ${programName}-$GOOS-$GOARCH ..."
go build -o ${programName}-$GOOS-$GOARCH "${currentDir}"/main.go
