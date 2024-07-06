package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"template/config"
	"template/internal/app"

	"git.astralnalog.ru/utils/aconf/v3"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// Чтение конфигураций из файла .env переданного в параметре запуска.
	var envPath string

	flag.StringVar(&envPath, "envPath", "", "путь к local.env")
	flag.Parse()

	if envPath != "" {
		err := aconf.PreloadEnvsFile(envPath)
		if err != nil {
			log.Fatalf("Ошибка загрузки конфигурационного файла: %s", err.Error())
		}
	}

	// Инициализация конфигурации.
	c := config.Config{}
	if err := aconf.Load(&c); err != nil {
		log.Fatalf("Ошибка инициализации конфигурации: %s", err.Error())
	}

	// Инициализация приложения.
	app, err := app.New(ctx, c)
	if err != nil {
		log.Fatalf("Ошибка инициализации приложения: %s", err.Error())
	}

	// Запуск приложения.
	err = app.Start(ctx)
	if err != nil {
		log.Fatalf("Ошибка запуска приложения: %s", err.Error())
	}
}
