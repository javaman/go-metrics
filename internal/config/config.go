package config

import (
	"flag"

	"github.com/caarlos0/env/v8"
)

type ServerConfiguration struct {
	Address string `env:"ADDRESS"`
}

type AgentConfiguration struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func ConfigureServer() *ServerConfiguration {
	config := &ServerConfiguration{}

	flag.StringVar(&config.Address, "a", "localhost:8080", "Адрес сервера")
	flag.Parse()

	env.Parse(config)

	return config
}

func ConfigureAgent() *AgentConfiguration {
	conf := &AgentConfiguration{}

	flag.StringVar(&conf.Address, "a", "localhost:8080", "Адрес сервера")
	flag.IntVar(&conf.ReportInterval, "r", 10, "Частота отправки на сервер")
	flag.IntVar(&conf.PollInterval, "p", 2, "Частота опроса метрик")
	flag.Parse()

	env.Parse(conf)

	return conf
}
