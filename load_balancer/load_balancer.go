package load_balancer

import (
	"context"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type LoadBalancer struct {
	servers  []*Backend
	strategy LoadBalancerStrategy
	port     int32
	current  int64
	mux      sync.RWMutex
	proxy    httputil.ReverseProxy
	cancel   context.CancelFunc
}

func NewLoadBalancer(urls []string, healthPath, healthInterval, healthTimeout string) (*LoadBalancer, error) {
	var servers []*Backend

	for _, rawUrl := range urls {
		if newUrl, err := url.Parse(rawUrl); err == nil {
			servers = append(servers, NewBackend(newUrl))
		}
	}

	if len(servers) == 0 {
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

	for _, server := range servers {
		go server.StartHealthCheck(ctx, interval, timeout, healthPath)
	}

	return &LoadBalancer{
		servers: servers,
		current: 0,
		mux:     sync.RWMutex{},
		cancel:  cancel,
	}, nil
}

func (l *LoadBalancer) Stop() {
	l.cancel()
}
