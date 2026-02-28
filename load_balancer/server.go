package load_balancer

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	url   *url.URL
	mux   sync.RWMutex
	Alive bool
}

func NewBackend(url *url.URL) *Backend {
	return &Backend{
		url:   url,
		mux:   sync.RWMutex{},
		Alive: false,
	}
}

func (s *Backend) StartHealthCheck(ctx context.Context, interval, timeout time.Duration, path string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.HealthCheck(timeout, path); err != nil {
				log.Printf("Health check %s: %v", s.url.String(), err)
			} else {
				log.Printf("Health check %s: alive=true", s.url.String())
			}
		case <-ctx.Done():
			log.Printf("Health check %s: stopping", s.url.String())
			return
		}
	}
}

func (s *Backend) HealthCheck(timeout time.Duration, path string) error {
	client := &http.Client{Timeout: timeout}
	healthUrl := s.url.String() + path
	resp, err := client.Get(healthUrl)
	if err != nil {
		s.SetAlive(false)
		return &HealthCheckError{ServerURL: s.url.String(), Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		s.SetAlive(false)
		return &HealthCheckError{ServerURL: s.url.String(), StatusCode: resp.StatusCode}
	}

	s.SetAlive(true)
	return nil
}

func (s *Backend) SetAlive(alive bool) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.Alive = alive
}

func (s *Backend) IsAlive() bool {
	s.mux.RLock()
	defer s.mux.RUnlock()
	return s.Alive
}
