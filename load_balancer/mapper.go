package load_balancer

import (
	"fmt"
	"strings"
)

func ToLoadBalancerStrategy(strategyStr string) (LoadBalancerStrategy, error) {
	formatted := strings.ToUpper(strings.ReplaceAll(strategyStr, "-", "_"))

	switch formatted {
	case "ROUND_ROBIN":
		return ROUND_ROBIN, nil
	default:
		return 0, &ConfigError{Field: "strategy", Value: strategyStr, Err: fmt.Errorf("unknown strategy")}
	}
}
