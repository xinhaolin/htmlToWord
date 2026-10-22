.PHONY: build test integration vet playground package docs
build:
	mkdir -p bin
	go build -trimpath -o bin/htmlword ./cmd/htmlword
test:
	go test -race ./...
	cd tools/playground && go test -race ./...
integration:
	HTMLWORD_INTEGRATION=1 go test -race ./...
vet:
	go vet ./...
	cd tools/playground && go vet ./...
playground:
	cd tools/playground && go run ./cmd/playground -addr 127.0.0.1:8787
package:
	python3 scripts/package-sdk.py
docs:
	python3 scripts/build-docs.py
	python3 scripts/build-docs.py --check
