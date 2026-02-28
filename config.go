package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server       ServerConfig       `yaml:"server"`
	LoadBalancer LoadBalancerConfig `yaml:"load_balancer"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type LoadBalancerConfig struct {
	Enabled   bool         `yaml:"enabled"`
	Strategy  string       `yaml:"strategy"`
	Health    HealthConfig `yaml:"health"`
	Resources []string     `yaml:"resources"`
}

type HealthConfig struct {
	Path     string `yaml:"path"`
	Interval string `yaml:"interval"`
	Timeout  string `yaml:"timeout"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
