package main

import (
	"fmt"
	"os"
)

type Config struct {
	Project      string
	Organization string
}

func ReadConfig() (*Config, error) {
	p := os.Getenv("PROJECT")
	if p == "" {
		return nil, fmt.Errorf("PROJECT env var is not set")
	}

	o := os.Getenv("ORGANIZATION")
	if o == "" {
		return nil, fmt.Errorf("ORGANIZATION env var is not set")
	}

	return &Config{Project: p, Organization: o}, nil
}
