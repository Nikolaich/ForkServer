#!/bin/bash

Vers=$1
Name=ForkServer

echo "package server" > server/version.go
echo "const Name, Vers = \"$Name\", \"$Vers\"" >> server/version.go

PLATFORMS=(
    'linux/arm'
    'linux/amd64'
    'linux/386'
    'linux/mipsle'
    'linux/mips'
    'darwin/arm64'
    'darwin/amd64'
    'windows/amd64'
    'windows/386'
)
for PLATFORM in "${PLATFORMS[@]}"; do
    o=${PLATFORM%/*}
    a=${PLATFORM#*/}
    e=""
    if [[ "$o" == "windows" ]]; then e=".exe"; fi
    f="distribs/$Name-$o-$a$e"
    echo -ne "> $f...\t"
    if [[ "$a" == "386" ]]; then
        GOOS=$o GOARCH=$a GO386=softfloat go build -o $f -ldflags="-s -w"
    else
        GOOS=$o GOARCH=$a go build -o $f -ldflags="-s -w"
    fi
    echo "done!"
done
