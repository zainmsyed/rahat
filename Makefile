SHELL := /bin/bash

GO ?= go
GOFMT ?= gofmt
NPM ?= npm

.PHONY: api db-setup test web web-install web-check web-build ci fmt fmt-check docker-build docker-smoke

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

docker-build:
	docker build -t rahat .

docker-smoke: docker-build
	@set -e; \
	container=$$(docker run -d --rm -p 8080:8080 -e TELEGRAM_BOT_TOKEN=test rahat); \
	trap 'docker stop $$container >/dev/null 2>&1 || true' EXIT; \
	for i in 1 2 3 4 5; do \
		if curl -fsS http://localhost:8080/healthz >/dev/null 2>&1; then \
			echo "smoke test passed"; \
			exit 0; \
		fi; \
		sleep 2; \
	done; \
	echo "smoke test failed"; \
	exit 1
