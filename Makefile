.PHONY: build test bench fuzz verify demo clean all default

default: all


GO ?= go
TMPDIR ?= $(shell pwd)/tmp

build:
	@mkdir -p tmp
	GOTMPDIR="$(TMPDIR)" $(GO) build ./src/...

test:
	@mkdir -p tmp
	GOTMPDIR="$(TMPDIR)" $(GO) test -v ./tests/port/...

bench:
	@mkdir -p tmp
	GOTMPDIR="$(TMPDIR)" $(GO) test -run '^$$' -bench=. -benchmem ./tests/port/...

fuzz:
	@mkdir -p tmp
	@echo "Running Cluster 9 Differential Fuzzer (requires Node.js v16+)..."
	GOTMPDIR="$(TMPDIR)" $(GO) run fuzz/harness/fuzz_cluster9.go

demo:
	@mkdir -p tmp
	GOTMPDIR="$(TMPDIR)" $(GO) run demo/main.go --auto

verify:
	@mkdir -p tmp
	GOTMPDIR="$(TMPDIR)" $(GO) test -v ./tests/port/... -run "TestProperty|TestStress|TestImmutability|TestCanonicalZero"

clean:
	$(GO) clean
	rm -rf tmp bin/*.exe

all: build test verify demo
