SHELL := /bin/bash

GO ?= go
NPM ?= npm

.PHONY: api db-setup web web-install web-build fmt

api:
	$(GO) run ./cmd/server

db-setup:
	$(GO) run ./cmd/server db:setup

web-install:
	cd web && $(NPM) install

web:
	cd web && $(NPM) run dev -- --host 0.0.0.0

web-build:
	cd web && $(NPM) run build

fmt:
	$(GO) fmt ./...
