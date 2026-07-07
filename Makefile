SHELL := /bin/bash

GO ?= go
NPM ?= npm

.PHONY: api db-setup test web web-install web-check web-build ci fmt

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

ci:
	$(GO) test ./...
	cd web && $(NPM) ci
	cd web && $(NPM) run check
	cd web && $(NPM) run build

fmt:
	$(GO) fmt ./...
