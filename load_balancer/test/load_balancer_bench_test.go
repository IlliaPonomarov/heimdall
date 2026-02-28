package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reverse-proxy/load_balancer"
	"testing"
)

func BenchmarkRoundRobinWithContainers(b *testing.B) {
	ctx := context.Background()

	backends, cleanup := startBackends(b, ctx, 3)
	defer cleanup()

	urls := make([]string, len(backends))
	for i, be := range backends {
		urls[i] = be.url
	}

	lb, err := load_balancer.NewLoadBalancer(urls, "round-robin", "/health", "30s", "2s")
	if err != nil {
		b.Fatalf("NewLoadBalancer: %v", err)
	}
	defer lb.Stop()

	for _, be := range lb.Backends() {
		be.SetAlive(true)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			lb.ServeHTTP(rec, req)
		}
	})
}

func BenchmarkRoundRobinStrategy(b *testing.B) {
	backends := make([]*load_balancer.Backend, 10)
	for i := range backends {
		u, _ := url.Parse("http://localhost:8080")
		backends[i] = load_balancer.NewBackend(u)
		backends[i].SetAlive(true)
	}

	strategy := &load_balancer.RoundRobinStrategy{}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			strategy.NextBackend(backends)
		}
	})
}

func BenchmarkServeHTTPInMemory(b *testing.B) {
	servers := make([]*httptest.Server, 3)
	urls := make([]string, 3)
	for i := range servers {
		servers[i] = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		}))
		urls[i] = servers[i].URL
	}
	defer func() {
		for _, s := range servers {
			s.Close()
		}
	}()

	lb, err := load_balancer.NewLoadBalancer(urls, "round-robin", "/", "30s", "2s")
	if err != nil {
		b.Fatalf("NewLoadBalancer: %v", err)
	}
	defer lb.Stop()

	for _, be := range lb.Backends() {
		be.SetAlive(true)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			lb.ServeHTTP(rec, req)
		}
	})
}
