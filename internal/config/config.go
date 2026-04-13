package config

import "time"

type Config struct {
	Replicas    int
	Namespace   string
	Image       string
	Concurrency int
	Timeout     time.Duration
	Label       string
}
