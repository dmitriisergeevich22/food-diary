package usecases

import (
	"context"
	"template/config"
	"template/pkg/metrics"

	"git.astralnalog.ru/edicore/contracts/fnstech"
	"github.com/xtls/xray-core/infra/conf"
	"golang.org/x/mod/sumdb/storage"
)

// Логика сервиса template.
type TemplateCore interface {
	// Логика сервиса.
	Logic(ctx context.Context) error

	// Проверка подключения к репозиторию.
	PingRepo(ctx context.Context) error
}

// Имплементация логики сервиса package-creator.
type templateCoreImpl struct {
	cfg  config.Config           // Конфигурация.
	repo packagerepo.PackageRepo // Репозиторий пакетов.
	// metricsCollector metrics.Collector               // Сборщик метрик.
}

// Инициализация сервиса.
func NewTemplateCore(cfg conf.Config, pkgRepo packagerepo.PackageRepo, brokerPublisher brokerpublisher.BrokerPublisher,
	storage storage.Storage, fnsHelper fnstech.PackageHelper, metricsCollector metrics.Collector) PackageCreatorCore {
	service := packageCreatorCoreImpl{
		cfg:              cfg,
		packageRepo:      pkgRepo,
		brokerPublisher:  brokerPublisher,
		storage:          storage,
		fnsHelper:        fnsHelper,
		metricsCollector: metricsCollector,
	}

	return &service
}

// Проверка подключения к репозиторию.
func (t *templateCoreImpl) PingRepo(ctx context.Context) error {
	return t.repo.Ping(ctx)
}
