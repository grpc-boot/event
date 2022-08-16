package config

import (
	"runtime"
)

type Config struct {
	App    App    `json:"app" yaml:"app"`
	Logger Logger `json:"x" yaml:"logger"`
}

func (c *Config) Format() *Config {
	if c.App.NumLoops < 1 {
		c.App.NumLoops = runtime.NumCPU()
	}

	if c.App.MaxIdleSeconds < 1 {
		c.App.MaxIdleSeconds = 60
	}

	return c
}
