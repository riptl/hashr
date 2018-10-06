#!/bin/sh
if [ -z "$1" ]
then
    echo "Usage: ./release.sh <version>"
    exit 1
fi
mkdir -p bin
GOOS=linux GOARCH=386 go build -o bin/hashr-$1-linux-i386
GOOS=linux GOARCH=amd64 go build -o bin/hashr-$1-linux-amd64
GOOS=linux GOARCH=arm go build -o bin/hashr-$1-linux-arm
GOOS=linux GOARCH=arm64 go build -o bin/hashr-$1-linux-arm64
GOOS=windows GOARCH=386 go build -o bin/hashr-$1-win32.exe
GOOS=windows GOARCH=amd64 go build -o bin/hashr-$1-win64.exe
GOOS=darwin GOARCH=amd64 go build -o bin/hashr-$1-macOS
