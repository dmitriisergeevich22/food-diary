package controller

import (
	"context"

	"github.com/gogo/status"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"

	"git.astralnalog.ru/utils/alogger"
)

// HealthCheck для k8s. k8s смотрит только на статус-код ответа. если он не 200 - начинает перезапускает поду.
// соответственно логика такая: если без чего-то сервис не может работать полностью (например без бд), то пусть падает
// если например недоступно api оператора - то частично функциональность сохраняется, поэтому 200.
func (gw *serverImpl) CheckHealth(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	g := errgroup.Group{}

	// Проверка репозитория.
	g.Go(func() error {
		err := gw.core.PingRepo(ctx)
		if err != nil {
			alogger.ErrorFromCtx(ctx, "ошибка при пинге БД приёмника", err, nil, false)
			return status.Errorf(codes.Internal, "ошибка при пинге БД приёмника: %s", err.Error())
		}

		return nil
	})

	// Проверка брокера издателя.
	g.Go(func() error {
		err := gw.core.PingBrokerPublisher(ctx)
		if err != nil {
			alogger.ErrorFromCtx(ctx, "ошибка при пинге NATS", err, nil, false)
			return status.Errorf(codes.Internal, "ошибка при пинге NATS: %s", err.Error())
		}

		return nil
	})

	// Проверка хранилища.
	g.Go(func() error {
		err := gw.core.PingStorage(ctx)
		if err != nil {
			alogger.ErrorFromCtx(ctx, "ошибка при пинге хранилища", err, nil, false)
			return status.Errorf(codes.Internal, "ошибка при пинге хранилища: %s", err.Error())
		}

		return nil
	})

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
