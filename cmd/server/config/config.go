package config

import (
	"flag"

	"github.com/caarlos0/env/v8"
)

type Configuration struct {
	Address string `env:"ADDRESS"`
}

func InitConfig() *Configuration {
	config := &Configuration{}

	flag.StringVar(&config.Address, "a", "localhost:8080", "Адрес сервера")
	flag.Parse()

	env.Parse(&config)

	return config
}
