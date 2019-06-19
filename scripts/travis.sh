#!/bin/bash

set -e

if [[ $# -eq 0 ]]; then
	git fetch --tags
	version=`git describe --tags`
else
    version="$1"
fi

if [[ "$version" == "" ]]; then
	echo "Error: Cannot determine version"
	exit 1
fi

mkdir -p build

GO111MODULE=on GOOS=linux GOARCH=arm GOARM=5 go build -o build/philter-linux-arm5 -ldflags "-X github.com/liamg/philter/internal/app/philter/version.Version=$version" ./cmd/philter 
GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o build/philter-linux-amd64 -ldflags "-X github.com/liamg/philter/internal/app/philter/version.Version=$version" ./cmd/philter

cat ./blacklist.txt > build/tmp.txt
cat ./custom.txt >> build/tmp.txt
cat build/tmp.txt | sort | uniq > build/blacklist.txt
rm build/tmp.txt
cp ./philter.service build/philter.service
