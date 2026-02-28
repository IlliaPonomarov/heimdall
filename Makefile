GO ?= $(or $(shell which go 2>/dev/null),$(lastword $(wildcard $(HOME)/sdk/go*/bin/go)),go)
GOIMPORTS ?= $(or $(shell which goimports 2>/dev/null),$(wildcard $(HOME)/go/bin/goimports),goimports)

.PHONY: fmt lint test build docker

fmt:
	$(GOIMPORTS) -w .

lint:
	@test -z "$$($(GOIMPORTS) -l .)" || { echo "Files not formatted:"; $(GOIMPORTS) -l .; exit 1; }

test:
	$(GO) test ./load_balancer/test/ -run='TestAllBackendsDown|TestHealthCheckMarksBackendDead|TestNewLoadBalancerInvalidConfig|TestUnknownStrategy' -count=1

build: lint
	$(GO) build ./...

docker:
	docker build -t heimdall-reverse-proxy .
