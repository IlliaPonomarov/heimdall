package load_balancer

import (
	"context"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type LoadBalancer struct {
	backends []*Backend
	strategy StrategyAlgorithm
	mux      sync.RWMutex
	cancel   context.CancelFunc
}

func NewLoadBalancer(
	urls []string,
	strategy, healthPath, healthInterval, healthTimeout string,
) (*LoadBalancer, error) {
	var backends []*Backend
	strategyAlgorithm, err := ToLoadBalancerStrategy(strategy)

	if err != nil {
		return nil, err
	}

	for _, rawUrl := range urls {
		if newUrl, err := url.Parse(rawUrl); err == nil {
			backends = append(backends, NewBackend(newUrl))
		}
	}

	if len(backends) == 0 {
		return nil, &NoBackendsError{}
	}

	interval, err := time.ParseDuration(healthInterval)
	if err != nil {
		return nil, &ConfigError{Field: "health.interval", Value: healthInterval, Err: err}
	}

	timeout, err := time.ParseDuration(healthTimeout)
	if err != nil {
		return nil, &ConfigError{Field: "health.timeout", Value: healthTimeout, Err: err}
	}

	ctx, cancel := context.WithCancel(context.Background())

	for _, server := range backends {
		go server.StartHealthCheck(ctx, interval, timeout, healthPath)
	}

	return &LoadBalancer{
		backends: backends,
		mux:      sync.RWMutex{},
		cancel:   cancel,
		strategy: strategyAlgorithm,
	}, nil
}

func (l *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) (int, error) {
	strategy := l.strategy
	backend, err := strategy.NextBackend(l.backends)

	if err != nil {
		return -1, err
	}

	backend.proxy.ServeHTTP(w, r)
	return 1, nil
}

func (l *LoadBalancer) Backends() []*Backend {
	l.mux.RLock()
	defer l.mux.RUnlock()
	return l.backends
}

func (l *LoadBalancer) Stop() {
	l.cancel()
}
