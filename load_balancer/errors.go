package load_balancer

import "fmt"

type HealthCheckError struct {
	ServerURL  string
	StatusCode int
	Err        error
}

func (e *HealthCheckError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("health check failed for %s: %v", e.ServerURL, e.Err)
	}
	return fmt.Sprintf("health check failed for %s: status %d", e.ServerURL, e.StatusCode)
}

func (e *HealthCheckError) Unwrap() error {
	return e.Err
}

type ConfigError struct {
	Field string
	Value string
	Err   error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("invalid config field %q value %q: %v", e.Field, e.Value, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

type NoBackendsError struct{}

func (e *NoBackendsError) Error() string {
	return "no valid backend URLs provided"
}
