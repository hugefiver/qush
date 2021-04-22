package config

import (
	"path/filepath"

	"github.com/go-ini/ini"
	"github.com/hugefiver/qush/configdir"
)

type ServerConfig struct {
	Addr string
	Port uint

	HostKeyPath string
	PasswdAuth  bool

	LogPath  string
	LogLevel string
}

func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Addr:        "0.0.0.0",
		Port:        22,
		HostKeyPath: filepath.Join(configdir.LocalConfig("qush"), "qush_host_key"),
		PasswdAuth:  true,
		LogLevel:    "Info",
		LogPath:     "",
	}
}

func LoadServerConfig(path string) (*ServerConfig, error) {
	i, err := ini.Load(path)
	if err != nil {
		// log.Println("Cannot load config:", err)
		return nil, err
	}

	config := DefaultServerConfig()

	// Section `Server`
	{
		s := i.Section("Server")
		config.Addr = s.Key("Addr").MustString("0.0.0.0")
		config.Port = s.Key("Port").MustUint(22)

		if k, err := s.GetKey("HostKeyPath"); err == nil {
			config.HostKeyPath = k.String()
		}
	}

	// Section `Auth`
	{
		s := i.Section("Auth")

		if k, err := s.GetKey("PasswdAuth"); err == nil {
			config.PasswdAuth = k.MustBool(config.PasswdAuth)
		}
	}

	// Section `Log`
	{
		s := i.Section("Log")

		config.LogLevel = s.Key("LogLevel").In(config.LogLevel, []string{"Debug", "Info", "Warning", "Error"})
		config.LogPath = s.Key("LogPath").MustString(config.LogPath)
	}

	return config, nil
}

type ClientConfig struct {
}

func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{}
}
