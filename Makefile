
build:
	go build ./cmd/philter/

test:
	go test -v ./...

run: build
	./philter
