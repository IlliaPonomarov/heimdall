package load_balancer

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	url   *url.URL
	mux   sync.RWMutex
	proxy *httputil.ReverseProxy
	Alive bool
}

func NewBackend(url *url.URL) *Backend {
	return &Backend{
		url:   url,
		mux:   sync.RWMutex{},
		proxy: httputil.NewSingleHostReverseProxy(url),
		Alive: false,
	}
}

func (b *Backend) URL() *url.URL {
	return b.url
}

func (b *Backend) StartHealthCheck(ctx context.Context, interval, timeout time.Duration, path string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := b.HealthCheck(timeout, path); err != nil {
				log.Printf("Health check %s: %v", b.url.String(), err)
			} else {
				log.Printf("Health check %s: alive=true", b.url.String())
			}
		case <-ctx.Done():
			log.Printf("Health check %s: stopping", b.url.String())
			return
		}
	}
}

func (b *Backend) HealthCheck(timeout time.Duration, path string) error {
	client := &http.Client{Timeout: timeout}
	healthUrl := b.url.String() + path
	resp, err := client.Get(healthUrl)
	if err != nil {
		b.SetAlive(false)
		return &HealthCheckError{ServerURL: b.url.String(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b.SetAlive(false)
		return &HealthCheckError{ServerURL: b.url.String(), StatusCode: resp.StatusCode}
	}

	b.SetAlive(true)
	return nil
}

func (b *Backend) SetAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.Alive
}
