GO ?= go
VERSION ?= 0.1.0
COMMIT ?= unknown
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BINARY ?= otter-install
LDFLAGS ?= -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)

.PHONY: build test vet check clean

build:
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

check: test vet

clean:
	rm -f $(BINARY)
