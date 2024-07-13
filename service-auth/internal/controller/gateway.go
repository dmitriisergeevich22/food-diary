// Package v1 implements routing paths. Each services in own file.
package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/valyala/fasthttp/reuseport"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	"git.astralnalog.ru/utils/alogger"
	"git.astralnalog.ru/edicore/common-libs/interceptorsgrpc"

	conf "git.astralnalog.ru/edicore/package-creator/config"
	"git.astralnalog.ru/edicore/package-creator/internal/usecases"
	"git.astralnalog.ru/edicore/package-creator/pkg/broker/brokerconsumer"
	"git.astralnalog.ru/edicore/package-creator/pkg/metrics"
	generic_proto "git.astralnalog.ru/edicore/package-creator/pkg/package-creator-service"
)

// Сервер package-creator.
type Server interface {
	Start(ctx context.Context) error
	gracefulStop()
}

// Имплементация сервера package-creator.
type serverImpl struct {
	// HTTP сервер.
	httpServer *fiber.App
	// gRPC сервер.
	grpcServer *grpc.Server
	// Брокер потребитель.
	brokerConsumer brokerconsumer.BrokerConsumer
	// Конфигурация.
	cfg conf.Config
	// Сборщик метрик.
	metricsCollector metrics.Collector
	// Логика сервиса package-creator.
	core usecases.PackageCreatorCore
	// Заглушка gRPC.
	generic_proto.UnimplementedPackageCreatorServer
}

func NewGatewayServer(ctx context.Context, cfg conf.Config, core usecases.PackageCreatorCore, mc metrics.Collector, broker brokerconsumer.BrokerConsumer) (Server, error) {
	gwmux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := generic_proto.RegisterPackageCreatorHandlerFromEndpoint(
		context.Background(),
		gwmux,
		cfg.Gateway.GRPC.Host+":"+cfg.Gateway.GRPC.Port,
		opts,
	)
	if err != nil {
		return nil, err
	}

	server := &serverImpl{
		cfg:              cfg,
		core:             core,
		metricsCollector: mc,
		brokerConsumer:   broker,
	}

	httpServer := newHTTPServer(gwmux, cfg.Gateway, mc)

	methodsExclude := map[string]struct{}{
		generic_proto.PackageCreator_CheckHealth_FullMethodName: {},
	}

	logFunc := func(ctx context.Context, msg string) {
		alogger.InfoFromCtx(ctx, msg, nil, nil, false)
	}

	interceptors := grpc.ChainUnaryInterceptor(
		interceptorsgrpc.TraceIDInterceptor(ctx),
		interceptorsgrpc.LoggerInterceptor(ctx, methodsExclude, logFunc),
		interceptorsgrpc.MetricInterceptor(ctx, methodsExclude, mc.AddAPIMethodUse),
		interceptorsgrpc.AuthInterceptor(ctx, methodsExclude, cfg.Gateway.AuthToken),
	)

	grpcServer := grpc.NewServer(
		interceptors,
	)

	reflection.Register(grpcServer)

	server.httpServer = httpServer
	server.grpcServer = grpcServer

	generic_proto.RegisterPackageCreatorServer(grpcServer, server)

	return server, nil
}

func (gw *serverImpl) Start(ctx context.Context) error {
	// Запуск брокера потребителя.
	err := gw.makeBrokerSubscribers(ctx)
	if err != nil {
		return fmt.Errorf("ошибка при запуске брокера потребителя для создания пакетов: %w", err)
	}

	alogger.InfoFromCtx(ctx, "успешный запуск брокера потребителя для создания пакетов", nil, nil, false)

	errChanCapacity := 3
	errChan := make(chan error, errChanCapacity)

	// Запуск GRPC сервера.
	go func() {
		adr := gw.cfg.Gateway.GRPC.Host + ":" + gw.cfg.Gateway.GRPC.Port

		grpcListener, err := reuseport.Listen("tcp4", adr)
		if err != nil {
			errChan <- fmt.Errorf("ошибка при запуске GRPC сервера: %w", err)
			return
		}

		alogger.InfoFromCtx(ctx, "запуск GRPC сервера на "+adr, nil, nil, false)
		defer alogger.InfoFromCtx(ctx, "GRPC сервер остановлен", nil, nil, false)

		err = gw.grpcServer.Serve(grpcListener)
		if err != nil {
			errChan <- fmt.Errorf("ошибка при запуске GRPC сервера: %w", err)
		}
	}()

	// Запуск HTTP сервера.
	go func() {
		adr := gw.cfg.Gateway.HTTP.Host + ":" + gw.cfg.Gateway.HTTP.Port

		alogger.InfoFromCtx(ctx, "запуск HTTP сервера на "+adr, nil, nil, false)
		defer alogger.InfoFromCtx(ctx, "HTTP сервер остановлен", nil, nil, false)

		err := gw.httpServer.Listen(adr)
		if err != nil {
			errChan <- fmt.Errorf("ошибка при запуске http сервера: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		gw.gracefulStop()
		return nil
	case err := <-errChan:
		return err
	}
}

func (gw *serverImpl) gracefulStop() {
	gw.grpcServer.GracefulStop()
	_ = gw.httpServer.Shutdown()

	gracefulStopWaitMillisecond := 100
	time.Sleep(time.Millisecond * time.Duration(gracefulStopWaitMillisecond))
}
