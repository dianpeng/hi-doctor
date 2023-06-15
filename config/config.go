package config

import (
	"gopkg.in/yaml.v3"

	"fmt"
	"io"
	"os"

	"github.com/dianpeng/hi-doctor/s14y"
)

// configuration of the whole hi-doctor
type Config struct {
	Assets           map[string]interface{} `yaml:"assets"`         // assets field
	ServerAddress    string                 `yaml:"server_address"` // server address of system
	ServiceDiscovery s14y.Config            `yaml:"service_discovery"`
}

func LoadConfig(data string) (*Config, error) {
	out := Config{
		Assets: make(map[string]interface{}),
	}
	if err := yaml.Unmarshal([]byte(data), &out); err != nil {
		return nil, fmt.Errorf("load_config: %s", err)
	}
	return &out, nil
}

func LoadConfigFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load_config: %s", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("load_config: %s", err)
	}
	return LoadConfig(string(data))
}
