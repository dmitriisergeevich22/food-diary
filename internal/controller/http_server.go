package controller

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	conf "git.astralnalog.ru/edicore/package-creator/config"
	"git.astralnalog.ru/edicore/package-creator/pkg/metrics"
)

func newHTTPServer(grpcGateway *runtime.ServeMux, cfg conf.Gateway, metricsCollector metrics.Collector) *fiber.App {
	server := fiber.New()

	server.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
		AllowMethods: "GET, POST, OPTIONS",
	}))

	server.Static("/swagger", cfg.PathToSwaggerDir)                           // Swagger.
	server.Static("/package-creator/swagger", cfg.PathToSwaggerDir)           // Swagger для локальной сборки.
	server.All("/metrics", adaptor.HTTPHandler(metricsCollector.ServeHTTP())) // Метрики.
	server.All("/*", adaptor.HTTPHandler(grpcGateway))                        // GRPC-gateway (все методы из proto-файла)

	return server
}
