GO = /Users/illiaponomarov/sdk/go1.26.0/bin/go
GOIMPORTS = /Users/illiaponomarov/go/bin/goimports

.PHONY: fmt lint test build

fmt:
	$(GOIMPORTS) -w .

lint:
	@test -z "$$($(GOIMPORTS) -l .)" || { echo "Files not formatted:"; $(GOIMPORTS) -l .; exit 1; }

test:
	$(GO) test ./load_balancer/test/ -run='TestAllBackendsDown|TestHealthCheckMarksBackendDead|TestNewLoadBalancerInvalidConfig|TestUnknownStrategy' -count=1

build: lint
	$(GO) build ./...
