package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Environment string

const (
	EnvLocal Environment = "local"
	EnvDev   Environment = "dev"
	EnvProd  Environment = "prod"
)

type Config struct {
	Env      Environment    `yaml:"env"      env:"ENV" env-default:"local"`
	Logger   LoggerConfig   `yaml:"logger"   env-prefix:"LOG_"`
	Service  ServiceConfig  `yaml:"service"  env-prefix:"SERVICE_"`
	Database DatabaseConfig `yaml:"database" env-prefix:"DB_"`
}

type LoggerConfig struct {
	Level string `yaml:"level" env:"LEVEL" env-default:"info"`
}

type ServiceConfig struct {
	Name string `yaml:"name" env:"NAME" env-default:"bookstores-api"`
	Host string `yaml:"host" env:"HOST" env-default:"0.0.0.0"`
	Port string `yaml:"port" env:"PORT" env-default:"8080"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"      env:"HOST"      env-default:"localhost"`
	Port     string `yaml:"port"      env:"PORT"      env-default:"5432"`
	User     string `yaml:"user"      env:"USER"      env-default:"postgres"`
	Password string `yaml:"password"  env:"PASSWORD"  env-default:"postgres"`
	DBName   string `yaml:"db_name"   env:"DB_NAME"   env-default:"bookstores"`
	SSLMode  string `yaml:"ssl_mode"  env:"SSL_MODE"  env-default:"disable"`
	MaxConns int    `yaml:"max_conns" env:"MAX_CONNS" env-default:"10"`
}

func Load(configPath string) (*Config, error) {
	cfg := &Config{}

	if configPath == "" {
		if err := cleanenv.ReadEnv(cfg); err != nil {
			return nil, fmt.Errorf("error reading config from environment variables: %w", err)
		}
		return cfg, nil
	}

	if err := cleanenv.ReadConfig(configPath, cfg); err != nil {
		return nil, fmt.Errorf("error reading config file '%s': %w", configPath, err)
	}

	return cfg, nil
}

func (dc *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dc.Host, dc.Port, dc.User, dc.Password, dc.DBName, dc.SSLMode)
}
