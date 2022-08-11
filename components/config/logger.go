package config

import (
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	Level     zapcore.Level `json:"level" yaml:"level"`
	DebugPath string        `json:"debugPath" yaml:"debugPath"`
	InfoPath  string        `json:"infoPath" yaml:"infoPath"`
	ErrorPath string        `json:"errorPath" yaml:"errorPath"`
}
