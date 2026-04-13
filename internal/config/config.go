package config

import "time"

type Config struct {
	DeploymentName string
	Namespace      string
	Image          string
	MinReplicas    int
	MaxReplicas    int
	Workers        int
	Cycles         int
	Dwell          time.Duration
	Timeout        time.Duration
	Cleanup        bool
}
