GO ?= go
SWIFT ?= swift
DOCKER_COMPOSE ?= docker compose

.PHONY: help phase1-check phase2-go-check phase3-go-check phase4-go-check phase5-go-check phase6-release-check phase7-hardening-check package-darwin-arm64 verify-release-artifact gofmt gofmt-check go-test go-vet go-race go-fuzz go-bench go-mod-verify govulncheck go-build swift-build swift-test test verify docker-clean

help:
	@printf '%s\n' \
		'Targets:' \
		'  phase1-check     Validate Phase 1 documentation artifacts in Docker' \
		'  phase2-go-check  Validate Go skeleton and docs in Docker' \
		'  phase3-go-check  Validate config, persistence, discovery, and suggestions in Docker' \
		'  phase4-go-check  Validate routing, route approval, and forwarding in Docker' \
		'  phase5-go-check  Validate setup, doctor, uninstall, and menu IPC support in Docker' \
		'  phase6-release-check Validate packaging templates and release checks in Docker' \
		'  phase7-hardening-check Validate hardening, fuzz, race, and release-candidate checks in Docker' \
		'  package-darwin-arm64 Build Portnado.app and release ZIP on macOS' \
		'  verify-release-artifact Validate local release ZIP and Cask checksum' \
		'  gofmt            Format Go code' \
		'  gofmt-check      Check Go formatting' \
		'  go-test          Run Go tests' \
		'  go-vet           Run go vet' \
		'  go-build         Build Go CLI and daemon' \
		'  swift-build      Build SwiftUI menu bar skeleton on macOS' \
		'  swift-test       Run Swift package tests on macOS' \
		'  test             Run Go and Swift tests' \
		'  verify           Run full available local verification'

phase1-check:
	$(DOCKER_COMPOSE) build phase1-checks
	$(DOCKER_COMPOSE) run --rm phase1-checks
	$(DOCKER_COMPOSE) down

phase2-go-check:
	$(DOCKER_COMPOSE) build phase2-go-checks
	$(DOCKER_COMPOSE) run --rm phase2-go-checks
	$(DOCKER_COMPOSE) down

phase3-go-check:
	$(DOCKER_COMPOSE) build phase3-go-checks
	$(DOCKER_COMPOSE) run --rm phase3-go-checks
	$(DOCKER_COMPOSE) down

phase4-go-check:
	$(DOCKER_COMPOSE) build phase4-go-checks
	$(DOCKER_COMPOSE) run --rm phase4-go-checks
	$(DOCKER_COMPOSE) down

phase5-go-check:
	$(DOCKER_COMPOSE) build phase5-go-checks
	$(DOCKER_COMPOSE) run --rm phase5-go-checks
	$(DOCKER_COMPOSE) down

phase6-release-check:
	$(DOCKER_COMPOSE) build phase6-release-checks
	$(DOCKER_COMPOSE) run --rm phase6-release-checks
	$(DOCKER_COMPOSE) down

phase7-hardening-check:
	$(DOCKER_COMPOSE) build phase7-hardening-checks
	$(DOCKER_COMPOSE) run --rm phase7-hardening-checks
	$(DOCKER_COMPOSE) down

package-darwin-arm64:
	scripts/package-darwin-arm64.sh

verify-release-artifact:
	scripts/verify-release-artifact.sh

gofmt:
	$(GO)fmt -w cmd internal pkg

gofmt-check:
	@test -z "$$($(GO)fmt -l cmd internal pkg)"

go-test:
	$(GO) test ./...

go-vet:
	$(GO) vet ./...

go-race:
	$(GO) test -race ./...

go-fuzz:
	$(GO) test -run=Fuzz -fuzz=FuzzParseRepository -fuzztime=10s ./internal/config
	$(GO) test -run=Fuzz -fuzz=FuzzValidateLocalhost -fuzztime=10s ./internal/config
	$(GO) test -run=Fuzz -fuzz=Fuzz -fuzztime=10s ./pkg/protocol

go-bench:
	$(GO) test -bench=. -benchmem ./internal/config ./internal/routing/http

go-mod-verify:
	$(GO) mod verify

govulncheck:
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest ./...

go-build:
	$(GO) build ./cmd/portnado
	$(GO) build ./cmd/portnado-daemon

swift-build:
	$(SWIFT) build --package-path apps/Portnado

swift-test:
	$(SWIFT) test --package-path apps/Portnado

test: go-test swift-test

verify: gofmt-check go-vet go-test swift-build swift-test phase7-hardening-check

docker-clean:
	$(DOCKER_COMPOSE) down
