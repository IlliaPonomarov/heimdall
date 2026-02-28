package test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reverse-proxy/load_balancer"
	"testing"
	"time"
)

func TestRoundRobinDistribution(t *testing.T) {
	ctx := context.Background()
	backends, cleanup := startBackends(t, ctx, 3)
	defer cleanup()

	urls := make([]string, len(backends))
	for i, b := range backends {
		urls[i] = b.url
	}

	lb, err := load_balancer.NewLoadBalancer(urls, "round-robin", "/health", "10s", "2s")
	if err != nil {
		t.Fatalf("NewLoadBalancer: %v", err)
	}
	defer lb.Stop()

	for _, b := range lb.Backends() {
		b.SetAlive(true)
	}

	counts := make(map[string]int)
	totalRequests := 12

	for i := 0; i < totalRequests; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		_, err := lb.ServeHTTP(rec, req)
		if err != nil {
			t.Fatalf("request %d: ServeHTTP error: %v", i, err)
		}

		body, _ := io.ReadAll(rec.Result().Body)
		counts[string(body)]++
	}

	if len(counts) != 3 {
		t.Errorf("expected requests distributed to 3 backends, got %d: %v", len(counts), counts)
	}
	for body, count := range counts {
		if count != 4 {
			t.Errorf("backend %q got %d requests, expected 4", body, count)
		}
	}
}

func TestSkipsUnhealthyBackend(t *testing.T) {
	ctx := context.Background()
	backends, cleanup := startBackends(t, ctx, 3)
	defer cleanup()

	urls := make([]string, len(backends))
	for i, b := range backends {
		urls[i] = b.url
	}

	lb, err := load_balancer.NewLoadBalancer(urls, "round-robin", "/health", "10s", "2s")
	if err != nil {
		t.Fatalf("NewLoadBalancer: %v", err)
	}
	defer lb.Stop()

	for _, b := range lb.Backends() {
		b.SetAlive(true)
	}
	lb.Backends()[1].SetAlive(false)

	counts := make(map[string]int)
	totalRequests := 10

	for i := 0; i < totalRequests; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		_, err := lb.ServeHTTP(rec, req)
		if err != nil {
			t.Fatalf("request %d: ServeHTTP error: %v", i, err)
		}

		body, _ := io.ReadAll(rec.Result().Body)
		counts[string(body)]++
	}

	if counts["backend-1"] != 0 {
		t.Errorf("dead backend-1 got %d requests, expected 0", counts["backend-1"])
	}

	if len(counts) != 2 {
		t.Errorf("expected 2 active backends, got %d: %v", len(counts), counts)
	}
}

func TestAllBackendsDown(t *testing.T) {
	lb, err := load_balancer.NewLoadBalancer(
		[]string{"http://localhost:9999", "http://localhost:9998"},
		"round-robin", "/health", "10s", "2s",
	)
	if err != nil {
		t.Fatalf("NewLoadBalancer: %v", err)
	}
	defer lb.Stop()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	_, err = lb.ServeHTTP(rec, req)
	if err == nil {
		t.Fatal("expected error when all backends are down, got nil")
	}

	var noHealthy *load_balancer.NoHealthyBackendsError
	if !errors.As(err, &noHealthy) {
		t.Errorf("expected NoHealthyBackendsError, got %T: %v", err, err)
	}
}

func TestHealthCheckMarksBackendAlive(t *testing.T) {
	ctx := context.Background()
	backends, cleanup := startBackends(t, ctx, 1)
	defer cleanup()

	parsedURL, err := url.Parse(backends[0].url)
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}

	backend := load_balancer.NewBackend(parsedURL)

	if backend.IsAlive() {
		t.Fatal("new backend should start as not alive")
	}

	err = backend.HealthCheck(5*time.Second, "/health")
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if !backend.IsAlive() {
		t.Error("backend should be alive after successful health check")
	}
}

func TestHealthCheckMarksBackendDead(t *testing.T) {
	parsedURL, _ := url.Parse("http://localhost:1")
	backend := load_balancer.NewBackend(parsedURL)
	backend.SetAlive(true)

	err := backend.HealthCheck(1*time.Second, "/health")
	if err == nil {
		t.Fatal("expected error for unreachable backend")
	}

	var hcErr *load_balancer.HealthCheckError
	if !errors.As(err, &hcErr) {
		t.Errorf("expected HealthCheckError, got %T: %v", err, err)
	}

	if backend.IsAlive() {
		t.Error("backend should be dead after failed health check")
	}
}

func TestNewLoadBalancerInvalidConfig(t *testing.T) {
	_, err := load_balancer.NewLoadBalancer([]string{"http://localhost:8080"}, "round-robin", "/health", "bad", "2s")
	var cfgErr *load_balancer.ConfigError
	if !errors.As(err, &cfgErr) {
		t.Errorf("expected ConfigError for bad interval, got %T: %v", err, err)
	}

	_, err = load_balancer.NewLoadBalancer([]string{"http://localhost:8080"}, "round-robin", "/health", "10s", "bad")
	if !errors.As(err, &cfgErr) {
		t.Errorf("expected ConfigError for bad timeout, got %T: %v", err, err)
	}

	_, err = load_balancer.NewLoadBalancer([]string{}, "round-robin", "/health", "10s", "2s")
	var noBackends *load_balancer.NoBackendsError
	if !errors.As(err, &noBackends) {
		t.Errorf("expected NoBackendsError, got %T: %v", err, err)
	}
}

func TestUnknownStrategy(t *testing.T) {
	_, err := load_balancer.NewLoadBalancer([]string{"http://localhost:8080"}, "least-connections", "/health", "10s", "2s")
	var cfgErr *load_balancer.ConfigError
	if !errors.As(err, &cfgErr) {
		t.Errorf("expected ConfigError for unknown strategy, got %T: %v", err, err)
	}
}
