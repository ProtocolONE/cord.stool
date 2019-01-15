package config

import (
	//"github.com/kelseyhightower/envconfig"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type DatabaseCfg struct {
	Host     string `yaml:"mongo_host"`
	Database string `yaml:"mongo_db"`
	User     string `yaml:"mongo_user"`
	Password string `yaml:"mongo_password"`
}

type ServiceCfg struct {
	HttpScheme      string `yaml:"http_scheme"`
	ServicePort     int    `yaml:"service_port"`
	PrivateKeyPath  string `yaml:"private_key_path"`
	PublicKeyPath   string `yaml:"public_key_path"`
	JwtExpDelta     int    `yaml:"jwt_expiration_delta"`
	StorageRootPath string `yaml:"storage_root_path"`
}

type Config struct {
	Database DatabaseCfg `yaml:"database"`
	Service  ServiceCfg  `yaml:"service"`
}

var cfg *Config

func Init() (*Config, error) {

	cfg = &Config{}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(filepath.Join(pwd, "service-config.yaml"))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	decoder.Decode(&cfg)

	/*if err = envconfig.Process("", cfg); err != nil {
		return nil, err
	}*/

	return cfg, nil
}

func Get() *Config {
	return cfg
}
