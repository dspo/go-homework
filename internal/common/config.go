package common

type ServerConfig struct {
	Listen ServerListenConfig `json:"listen" yaml:"listen"`
}

type ServerListenConfig struct {
	Host string `json:"host" yaml:"host" mapstructure:"host"`
	Port int    `json:"port" yaml:"port" mapstructure:"port"`
}
