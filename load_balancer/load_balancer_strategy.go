package load_balancer

import "sync/atomic"

type StrategyAlgorithm interface {
	NextBackend(backends []*Backend) (*Backend, error)
}

type RoundRobinStrategy struct {
	current int64
}

func (r *RoundRobinStrategy) NextBackend(backends []*Backend) (*Backend, error) {
	if len(backends) == 0 {
		return nil, &NoBackendsError{}
	}
	total := len(backends)
	next := int(atomic.AddInt64(&r.current, 1))

	for i := 0; i < total; i++ {
		idx := (i + next) % total
		backend := backends[idx]
		if backend.IsAlive() {
			return backend, nil
		}
	}
	return nil, &NoHealthyBackendsError{}
}
