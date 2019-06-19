#!/bin/bash

set -e

version='branch'

if [[ ! -z "$TRAVIS_TAG" ]]; then
	version="${TRAVIS_TAG}"
fi

mkdir -p build

GO111MODULE=on GOOS=linux GOARCH=arm GOARM=5 go build -o build/philter-linux-arm5 -ldflags "-X github.com/liamg/philter/internal/app/philter/version.Version=$version" ./cmd/philter 
GO111MODULE=on GOOS=linux GOARCH=amd64 go build -o build/philter-linux-amd64 -ldflags "-X github.com/liamg/philter/internal/app/philter/version.Version=$version" ./cmd/philter

cat ./blacklist.txt > build/tmp.txt
cat ./custom.txt >> build/tmp.txt
cat build/tmp.txt | sort | uniq > build/blacklist.txt
rm build/tmp.txt
cp ./philter.service build/philter.service
