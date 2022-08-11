package config

type App struct {
	Addr        string `json:"addr" yaml:"addr"`
	PprofAddr   string `json:"pprofAddr" yaml:"pprofAddr"`
	MaxWorkers  uint32 `json:"maxWorkers" yaml:"maxWorkers"`
	IoTimeoutMs uint32 `json:"ioTimeoutMs" yaml:"ioTimeoutMs"`
}
