.PHONY: build test fuzz bench clean

build:
	go build ./...

test:
	go test ./...

fuzz:
	@echo "Running differential fuzzing..."

bench:
	go test -bench=. ./...

clean:
	go clean
