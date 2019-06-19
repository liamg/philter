
build:
	GO111MODULE=on go build ./cmd/philter/

test:
	GO111MODULE=on go test -v ./...

run: build
	./philter

travis:
	./scripts/travis.sh
	
pi: travis
	./scripts/pi.sh
