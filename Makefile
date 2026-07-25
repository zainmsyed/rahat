SHELL := /bin/bash

GO ?= go
GOFMT ?= gofmt
NPM ?= npm

.PHONY: api db-setup test web web-install web-check web-build ci fmt fmt-check

api:
	$(GO) run ./cmd/server

db-setup:
	$(GO) run ./cmd/server db:setup

web-install:
	cd web && $(NPM) install

test:
	$(GO) test ./...

web:
	cd web && $(NPM) run dev -- --host 0.0.0.0

web-check:
	cd web && $(NPM) run check

web-build:
	cd web && $(NPM) run build

ci: fmt-check
	$(GO) test ./...
	cd web && $(NPM) ci
	cd web && $(NPM) run check
	cd web && $(NPM) run build

fmt:
	$(GOFMT) -w $$(find . -type f -name '*.go' -not -path './vendor/*')

fmt-check:
	@test -z "$$($(GOFMT) -l $$(find . -type f -name '*.go' -not -path './vendor/*'))" || { \
		echo "Go files need formatting; run 'make fmt'"; \
		$(GOFMT) -l $$(find . -type f -name '*.go' -not -path './vendor/*'); \
		exit 1; \
	}
