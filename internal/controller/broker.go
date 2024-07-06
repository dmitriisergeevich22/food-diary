package controller

import (
	"context"
	"fmt"
	"time"

	"git.astralnalog.ru/utils/alogger"
	anats "git.astralnalog.ru/utils/anats"

	"git.astralnalog.ru/edicore/package-creator/internal/entity"
)

// Создание подписчиков на очереди.
func (gw *serverImpl) makeBrokerSubscribers(ctx context.Context) error {
	// Очередь создания пакета и отправки его в package-sender.
	optCreatePackageAndSend := anats.SubscribeOptions{
		Workers:           gw.cfg.Queue.Workers,
		MaxDeliver:        gw.cfg.Queue.MaxDeliver,
		AckWaitSeconds:    gw.cfg.Queue.AckWaitSeconds,
		NakTimeoutSeconds: gw.cfg.Queue.NakTimeoutSeconds,
		MaxAckPending:     gw.cfg.Queue.MaxAckPending,
	}

	err := gw.brokerConsumer.SubscribeToCreatePackageAndSend(ctx, gw.handlerForCreatePackageAndSend, optCreatePackageAndSend)
	if err != nil {
		return fmt.Errorf("ошибка создания подписки для очереди приёма документов - %s", err.Error())
	}

	return nil
}

// Обработчик события создания пакета и отправки в сервис package-sender.
func (gw *serverImpl) handlerForCreatePackageAndSend(ctx context.Context, desc entity.PackageDescription, retry int) anats.MessageResultEnum {
	start := time.Now()
	defer func() { gw.metricsCollector.AddPkgProcessTime(time.Since(start)) }()

	if retry == 1 {
		gw.metricsCollector.IncPkgCounter()
	}

	alogger.DebugFromCtx(ctx, fmt.Sprintf("начало создания пакета %s из брокера", desc.PackageName), nil, nil, false)

	aerr := gw.core.CreatePackageAndSend(ctx, desc)
	if aerr != nil {
		alogger.ErrorFromCtx(ctx, fmt.Sprintf("ошибка обработки пакета %s: %s", desc.PackageName, aerr.DeveloperMessage()), aerr, nil, false)

		if aerr.IsCritical() {
			gw.metricsCollector.AddCriticalError(aerr.DeveloperMessage())
			return anats.MessageResultEnumFatalError
		}

		gw.metricsCollector.AddTempError(aerr.DeveloperMessage())

		return anats.MessageResultEnumTempError
	}

	alogger.InfoFromCtx(ctx, fmt.Sprintf("окончание обработки создания пакета %s из брокера. время - %.3fs", desc.PackageName, time.Since(start).Seconds()), nil, nil, false)

	return anats.MessageResultEnumSuccess
}
