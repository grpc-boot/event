package config

type App struct {
	Addr           string `json:"addr" yaml:"addr"`
	PprofAddr      string `json:"pprofAddr" yaml:"pprofAddr"`
	NumLoops       int    `json:"numLoops" yaml:"numLoops"`
	MaxIdleSeconds uint32 `json:"maxIdleSeconds" yaml:"maxIdleSeconds"`
}
