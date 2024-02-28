package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"log"
	"os"
	"time"
)

type Config struct {
	Env            string     `yaml:"env" env-default:"development"`
	ManticoreIndex string     `yaml:"manticore_index"`
	StartURLs      []StartURL `yaml:"start_urls"`
	Parser         `yaml:"parser"`
}

type StartURL struct {
	Lang string `yaml:"lang"`
	Url  string `yaml:"url"`
}

type Parser struct {
	ResourceID int            `yaml:"resource_id" env-default:"1"`
	PageCount  int            `yaml:"page_count" env-default:"1"`
	OutputPath string         `yaml:"output_path" env-default:"./data"`
	ParseDelay *time.Duration `yaml:"parse_delay" env-default:"5s"`
}

func MustLoad() *Config {
	// Получаем путь до конфиг-файла из env-переменной CONFIG_PATH
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	// Проверяем существование конфиг-файла
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("error opening config file: %s", err)
	}

	var cfg Config

	// Читаем конфиг-файл и заполняем нашу структуру
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	return &cfg
}
