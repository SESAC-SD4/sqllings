.PHONY: build test clean

build:
	bash build.sh

test:
	go test ./... -timeout 120s

clean:
	rm -f sqllings sqllings.exe sqllings-darwin-arm64 sqllings-darwin-amd64
