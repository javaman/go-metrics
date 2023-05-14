package config

import (
	"flag"

	"github.com/caarlos0/env/v8"
)

type Configuration struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
}

func Configure() *Configuration {
	conf := &Configuration{}

	flag.StringVar(&conf.Address, "a", "localhost:8080", "Адрес сервера")
	flag.IntVar(&conf.ReportInterval, "r", 10, "Частота отправки на сервер")
	flag.IntVar(&conf.PollInterval, "p", 2, "Частота опроса метрик")
	flag.Parse()

	env.Parse(conf)

	return conf
}
