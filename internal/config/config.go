package config

import (
	"flag"

	"github.com/caarlos0/env/v8"
)

type ServerConfiguration struct {
	Address         string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DBDsn           string `env:"DATABASE_DSN"`
	Key             string `env:"KEY"`
}

type AgentConfiguration struct {
	Address        string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	Key            string `env:"KEY"`
}

func ConfigureServer() *ServerConfiguration {
	conf := &ServerConfiguration{}

	flag.StringVar(&conf.Address, "a", "localhost:8080", "Адрес сервера")
	flag.IntVar(&conf.StoreInterval, "i", 300, "Интервал сохранения на диск. 0 - синхронно")
	flag.StringVar(&conf.FileStoragePath, "f", "/tmp/metrics-db.json", "Файл, где сохраняются метрики")
	flag.BoolVar(&conf.Restore, "r", false, "Загрузить ли ранее сохраненные значения")
	flag.StringVar(&conf.DBDsn, "d", "", "Подключение к БД")
	flag.StringVar(&conf.Key, "k", "", "Ключ")
	flag.Parse()

	env.Parse(conf)

	return conf
}

func ConfigureAgent() *AgentConfiguration {
	conf := &AgentConfiguration{}

	flag.StringVar(&conf.Address, "a", "localhost:8080", "Адрес сервера")
	flag.IntVar(&conf.ReportInterval, "r", 10, "Частота отправки на сервер")
	flag.IntVar(&conf.PollInterval, "p", 2, "Частота опроса метрик")
	flag.StringVar(&conf.Key, "k", "", "Ключ")
	flag.Parse()

	env.Parse(conf)

	return conf
}
