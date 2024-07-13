package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/status"
	"google.golang.org/grpc/codes"

	"git.astralnalog.ru/utils/alogger"

	"git.astralnalog.ru/edicore/package-creator/internal/entity"
	package_creator "git.astralnalog.ru/edicore/package-creator/pkg/package-creator-service"
)

// TODO возможно вместо status.Errorf возвращать ответ с описанием ошибки и статусом.
func (gw *serverImpl) CreatePackage(ctx context.Context, in *package_creator.CreatePackageRequest) (*package_creator.CreatePackageResponse, error) {
	start := time.Now()

	gw.metricsCollector.IncPkgCounter()
	defer func() { gw.metricsCollector.AddPkgProcessTime(time.Since(start)) }()

	alogger.DebugFromCtx(ctx, fmt.Sprintf("начало создания пакета %s из api", in.GetPackageName()), nil, nil, false)

	// Валидация
	if in.GetPackageType() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "не передан тип пакета")
	}

	if in.GetPackageName() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "не передано имя пакета")
	}

	if in.GetDestinationURL() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "не передан URL назначения")
	}

	if in.GetReceiverOperatorID() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "не передан код оператора получателя")
	}

	if in.GetSenderOperatorID() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "не передан код получателя получателя")
	}

	if len(in.GetDescription()) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "не передано описание пакета")
	}

	// Трансформация в entity.PackageDescription.
	pkgType, err := entity.StringToPackageType(in.PackageType)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка преобразования типа пакета: %s", err.Error())
	}

	desc := entity.PackageDescription{
		PackageType:        pkgType,
		PackageName:        in.GetPackageName(),
		DestinationURL:     in.GetDestinationURL(),
		ReceiverIsHub:      in.GetReceiverIsHub(),
		ReceiverOperatorID: in.GetReceiverOperatorID(),
		SenderOperatorID:   in.GetSenderOperatorID(),
		Description:        in.GetDescription(),
	}

	// Запуск создания пакета и отправки в package-sender.
	aerr := gw.core.CreatePackageAndSend(ctx, desc)
	if aerr != nil {
		alogger.ErrorFromCtx(ctx, fmt.Sprintf("ошибка обработки пакета %s из api: %s", desc.PackageName, aerr.DeveloperMessage()), aerr, nil, false)

		if aerr.IsCritical() {
			gw.metricsCollector.AddCriticalError(aerr.DeveloperMessage())
		} else {
			gw.metricsCollector.AddTempError(aerr.DeveloperMessage())
		}

		return nil, status.Errorf(codes.Internal, "ошибка обработки пакета %s: %s", desc.PackageName, aerr.Error())
	}

	return &package_creator.CreatePackageResponse{Status: true}, nil
}
