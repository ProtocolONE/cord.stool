package config

import (
	"github.com/kelseyhightower/envconfig"
)

type DatabaseEnv struct {
	Host     string `envconfig:"MONGO_HOST"`
	Database string `envconfig:"MONGO_DB"`
	User     string `envconfig:"MONGO_USER"`
	Password string `envconfig:"MONGO_PASSWORD"`
}

type ServiceEnv struct {
	HttpScheme     string `envconfig:"HTTP_SCHEME"`
	ServicePort    int    `envconfig:"SERVICE_PORT"`
	PrivateKeyPath string `envconfig:"PRIVATE_KEY_PATH"`
	PublicKeyPath  string `envconfig:"PUBLIC_KEY_PATH"`
	JwtExpDelta    int    `envconfig:"JWT_EXPIRATION_DELTA"`
}

type Config struct {
	DatabaseEnv
	ServiceEnv
}

var cfg *Config

func Init() (error, *Config) {
	var err error
	cfg = &Config{}
	if err = envconfig.Process("", cfg); err != nil {
		return err, nil
	}
	return nil, cfg
}

func Get() *Config {
	return cfg
}
