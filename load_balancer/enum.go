package load_balancer

type LoadBalancerStrategy uint16

const (
	ROUND_ROBIN LoadBalancerStrategy = iota
)
