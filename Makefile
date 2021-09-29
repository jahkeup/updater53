GOLANGCI_LINT=golangci-lint
GO=go
V=

build: T=.
build:
	$(GO) build $(V) -mod=readonly $(T)

test: T=./...
test:
	$(GO) test $(V) -race -mod=readonly $(T)

lint: T=./...
lint:
	$(GOLANGCI_LINT) run $(T)
