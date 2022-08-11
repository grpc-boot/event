package config

type Config struct {
	App    App    `json:"app" yaml:"app"`
	Logger Logger `json:"logger" yaml:"logger"`
}
