package app

import (
	"context"
	"fmt"
	"template/config"
	"template/internal/controller"
	"template/internal/usecases"

	"git.astralnalog.ru/utils/postgresdb"
)

// Приложение.
type Application struct {
	Server controller.Server
}

func New(ctx context.Context, conf config.Config) (*Application, error) {
	// Настройка логгера. (опционально)

	// Инициализация сборщика метрик. (опционально)

	// Подключение к БД.
	conn, err := postgresdb.New(ctx, conf.PGWriterConn, conf.PGReaderConn)
	if err != nil {
		return nil, err
	}

	// Подключение репозитория.
	repo, err := repo.NewPackageRepo(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("ошибка при инициализации репозитория: %w", err)
	}

	if err := repo.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ошибка при подключении к репозиторию: %w", err)
	}

	// Основной бизнес-сервис.
	templateCore := usecases.NewPackageCreatorCore(conf, repo, brokerPublisher, st, fnsHelper, metricsCollector)

	// Создание сервера package-creator.
	server, err := controller.NewGatewayServer(ctx, conf, packageCreatorCore, metricsCollector, brokerConsumer)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании gateway сервиса: %s", err.Error())
	}

	return &Application{
		Server: server,
	}, nil
}

func (a *Application) Start(ctx context.Context) error {
	err := a.Server.Start(ctx)
	if err != nil {
		return fmt.Errorf("ошибка при работе gateway сервиса: %s", err.Error())
	}

	return nil
}
