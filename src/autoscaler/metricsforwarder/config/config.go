package config

import (
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	ServerPort            string `yaml:"server_port"`
	LogLevel              string `yaml:"log_level"`
	MetronAddress         string `yaml:"metron_address"`
	LoggregatorCaCertFile string `yaml:"loggregator_ca_cert_path"`
	LoggregatorCertFile   string `yaml:"loggregator_cert_path"`
	LoggregatorKeyFile    string `yaml:"loggregator_key_path"`
}

const (
	defaultLogLevel      = "info"
	defaultMetronAddress = "127.0.0.1:3457"
	defaultServerPort    = "7654"
)

func LoadConfig(reader io.Reader) (*Config, error) {
	conf := &Config{
		ServerPort:    defaultServerPort,
		LogLevel:      defaultLogLevel,
		MetronAddress: defaultMetronAddress,
	}

	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(bytes, conf)
	if err != nil {
		return nil, err
	}

	return conf, nil
}
