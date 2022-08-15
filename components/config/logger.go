package config

type Logger struct {
	Level     int8   `json:"level" yaml:"level"`
	DebugPath string `json:"debugPath" yaml:"debugPath"`
	InfoPath  string `json:"infoPath" yaml:"infoPath"`
	ErrorPath string `json:"errorPath" yaml:"errorPath"`
}
