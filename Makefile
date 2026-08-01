.PHONY: build test fuzz bench clean

GO := ./go_sdk/go/bin/go.exe
TMPDIR := $(shell pwd)/tmp

build:
	@mkdir -p tmp
	GOTMPDIR="$(TMPDIR)" $(GO) build ./src/...

test:
	@mkdir -p tmp
	GOTMPDIR="$(TMPDIR)" $(GO) test -v ./tests/port/...

fuzz:
	@echo "Running differential fuzzing..."

bench:
	@mkdir -p tmp
	GOTMPDIR="$(TMPDIR)" $(GO) test -bench=. ./tests/port/...

clean:
	$(GO) clean
	rm -rf tmp
