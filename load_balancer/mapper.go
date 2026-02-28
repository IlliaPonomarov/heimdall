package load_balancer

import (
	"fmt"
	"strings"
)

func ToLoadBalancerStrategy(strategyStr string) (StrategyAlgorithm, error) {
	formatted := strings.ToUpper(strings.ReplaceAll(strategyStr, "-", "_"))

	switch formatted {
	case "ROUND_ROBIN":
		return &RoundRobinStrategy{}, nil
	default:
		return nil, &ConfigError{Field: "strategy", Value: strategyStr, Err: fmt.Errorf("unknown strategy")}
	}
}
