package usecases

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	fnstech_entity "git.astralnalog.ru/edicore/contracts/fnstech/entity"
	aerror "git.astralnalog.ru/utils/aerror"
	"git.astralnalog.ru/utils/alogger"

	"git.astralnalog.ru/edicore/package-creator/internal/entity"
)

func (p *packageCreatorCoreImpl) CreatePackageAndSend(ctx context.Context, desc entity.PackageDescription) aerror.AError {
	p.metricsCollector.IncCurrentPkgInWork()
	defer p.metricsCollector.DecCurrentPkgInWork()

	// Изначально считаем что пакет пришел повторно.
	event := entity.PackageEventTypeGotAgain

	// Проверка наличия пакета.
	pkg, err := p.packageRepo.SelectPackageByName(ctx, desc.PackageName)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return aerror.New(ctx, entity.SelectPkgErrorID, err, "Ошибка проверки наличия пакета %s в БД: %s", desc.PackageName, err.Error())
	}

	// Если пакет не найден - создаем новый.
	if pkg == nil {
		event = entity.PackageEventTypeCreated

		pkg, err = p.packageRepo.InsertPackage(ctx, desc.PackageType, desc.PackageName, desc.DestinationURL, desc.ReceiverIsHub, desc.ReceiverOperatorID, desc.SenderOperatorID)
		if err != nil {
			return aerror.New(ctx, entity.InsertPkgErrorID, err, "Ошибка вставки пакета в БД: %s", err.Error())
		}
	}

	// Обновление данных пакета.
	pkg.Type = desc.PackageType
	pkg.DestinationURL = desc.DestinationURL

	// Создание события получения пакета. (EventTypePackageGotAgain || EventTypePackageCreated)
	err = p.packageRepo.AddNewEvent(pkg.ID, event, pkg.Name)
	if err != nil {
		return aerror.New(ctx, entity.InsertEventErrorID, err, "Ошибка вставки события получения пакета в БД: %s", err.Error())
	}

	// Пакет уже был обработан.
	if pkg.Status == entity.PackageStatusSuccess {
		alogger.InfoFromCtx(ctx, fmt.Sprintf("Пакет %s уже был обработан.", pkg.Name), nil, nil, false)
		return p.sendPackage(ctx, pkg)
	}

	// Упаковка пакета в cms.
	alogger.DebugFromCtx(ctx, fmt.Sprintf("Упаковка пакета %s в cms.", pkg.Name), nil, nil, false)
	cms, aerr := p.createCMS(ctx, desc)
	// Сохраним результат обработки в БД.
	aerr = p.saveResultForCreatePackage(ctx, pkg, aerr)
	if aerr != nil {
		return aerr
	}

	// Сохранение пакета в хранилище.
	alogger.DebugFromCtx(ctx, fmt.Sprintf("Сохранение пакета %s в хранилище.", pkg.Name), nil, nil, false)

	startSaveTime := time.Now()

	err = p.storage.SaveFile(ctx, pkg.Name, cms)
	if err != nil {
		return aerror.New(ctx, entity.SaveStorageErrorID, err, "Ошибка при сохранении пакета %s в хранилище.", pkg.Name)
	}

	p.metricsCollector.AddSaveStorageTime(time.Since(startSaveTime))

	// Отправка пакета в сервис package-sender.
	return p.sendPackage(ctx, pkg)
}

// Создание CMS.
func (p *packageCreatorCoreImpl) createCMS(ctx context.Context, desc entity.PackageDescription) ([]byte, aerror.AError) {
	var (
		cms  []byte
		aerr aerror.AError
	)

	// Валидация данных.
	if desc.Description == nil {
		return nil, aerror.NewCritical(ctx, entity.EmptyBodyErrorID, nil, "Получено пустое описание пакета.")
	}

	switch desc.PackageType {
	case entity.PackageTypeMessages:
		cms, aerr = p.packMessagesToCMS(ctx, desc.Description)
	case entity.PackageTypeInvitation:
		cms, aerr = p.packInvitationToCMS(ctx, desc.Description)
	case entity.PackageTypeTechnicalReceipt:
		cms, aerr = p.packTechnicalReceiptToCMS(ctx, desc.Description)
	default:
		return nil, aerror.New(ctx, entity.UnknownPkgTypeErrorID, nil, "Неизвестный тип пакета.")
	}

	if aerr != nil {
		return nil, aerr
	}

	return cms, nil
}

// Упаковка сообщений в cms.
func (p *packageCreatorCoreImpl) packMessagesToCMS(ctx context.Context, desc []byte) ([]byte, aerror.AError) {
	var msgs []entity.Message

	err := json.Unmarshal(desc, &msgs)
	if err != nil {
		return nil, aerror.New(ctx, entity.UnmarshalErrorID, err, "Ошибка при десериализации описания пакета: %s", err.Error())
	}

	// Подготовка логических сообщений.
	var logicalMessages []fnstech_entity.LogicalMessage

	for i := range msgs {
		// Подготовка списка файлов сообщения.
		// Файл description не нужен т.к. будет автоматически создан при упаковке в fnstech.
		files := make(map[string][]byte, len(msgs[i].Files))

		// Получение файлов из хранилища.
		for _, fileName := range msgs[i].Files {
			data, err := p.storage.GetFile(ctx, fileName)
			if err != nil {
				return nil, aerror.New(ctx, entity.GetStorageErrorID, err, "Ошибка при получении файла %s из хранилища: %s", fileName, err.Error())
			}

			files[fileName] = data
		}

		// Добавление логического сообщения.
		logicalMessages = append(logicalMessages, fnstech_entity.LogicalMessage{
			ID:    msgs[i].MessageID,
			Files: files,
			Desc: fnstech_entity.Description{
				Message: &msgs[i].Description,
			},
		})
	}

	// Упаковка сообщений в cms.
	cms, err := p.fnsHelper.PackLogicalMessages(ctx, logicalMessages)
	if err != nil {
		return nil, aerror.New(ctx, entity.PackCMSErrorID, err, "Ошибка при упаковке сообщений.")
	}

	return cms, nil
}

// Упаковка приглашения в cms.
func (p *packageCreatorCoreImpl) packInvitationToCMS(ctx context.Context, desc []byte) ([]byte, aerror.AError) {
	var inv entity.Invitation

	err := json.Unmarshal(desc, &inv)
	if err != nil {
		return nil, aerror.New(ctx, entity.UnmarshalErrorID, err, "Ошибка при десериализации описания пакета: %s", err.Error())
	}

	// Подготовка списка файлов сообщения.
	// Файл description не нужен т.к. будет автоматически создан при упаковке в fnstech.
	files := make(map[string][]byte, len(inv.Files))

	// Получение файлов из хранилища.
	for _, fileName := range inv.Files {
		data, err := p.storage.GetFile(ctx, fileName)
		if err != nil {
			return nil, aerror.New(ctx, entity.GetStorageErrorID, err, "Ошибка при получении файла %s из хранилища: %s", fileName, err.Error())
		}

		files[fileName] = data
	}

	// Подготовка логического сообщения.
	logicalMessage := fnstech_entity.LogicalMessage{
		ID:    inv.MessageID,
		Files: files,
		Desc: fnstech_entity.Description{
			Invitation: &inv.Description,
		},
	}

	// Упаковка приглашения в cms.
	cms, err := p.fnsHelper.PackLogicalMessages(ctx, []fnstech_entity.LogicalMessage{logicalMessage})
	if err != nil {
		return nil, aerror.New(ctx, entity.PackCMSErrorID, err, "Ошибка при упаковке сообщений.")
	}

	return cms, nil
}

// Упаковка технологической квитанции в cms.
func (p *packageCreatorCoreImpl) packTechnicalReceiptToCMS(ctx context.Context, desc []byte) ([]byte, aerror.AError) {
	var tr entity.TechnicalReceipt

	err := json.Unmarshal(desc, &tr)
	if err != nil {
		return nil, aerror.New(ctx, entity.UnmarshalErrorID, err, "Ошибка при десериализации описания пакета: %s", err.Error())
	}

	// Подготовка списка файлов квитанции.
	// Файл description не нужен т.к. будет автоматически создан при упаковке в fnstech.
	files := make(map[string][]byte, len(tr.Files))

	// Получение файлов из хранилища.
	for _, fileName := range tr.Files {
		data, err := p.storage.GetFile(ctx, fileName)
		if err != nil {
			return nil, aerror.New(ctx, entity.GetStorageErrorID, err, "Ошибка при получении файла %s из хранилища: %s", fileName, err.Error())
		}

		files[fileName] = data
	}

	// Добавление полученных файлов в квитанцию.
	tr.Description.FileCatalog = files

	// Упаковка технологической квитанции в cms.
	cms, err := p.fnsHelper.PackTechnicalReceipt(ctx, tr.Description)
	if err != nil {
		return nil, aerror.New(ctx, entity.PackCMSErrorID, err, "Ошибка при упаковке технологической квитанции: %s", err.Error())
	}

	return cms, nil
}

// Сохранение результата создания пакета.
func (p *packageCreatorCoreImpl) saveResultForCreatePackage(ctx context.Context, pkg *entity.Package, res aerror.AError) aerror.AError {
	var (
		event     entity.PackageEventType
		eventDesc string
	)

	if res == nil {
		pkg.ErrorCode = ""
		pkg.ErrorText = ""
		event = entity.PackageEventTypeSuccess
		eventDesc = pkg.Name
		pkg.Status = entity.PackageStatusSuccess
	} else {
		pkg.ErrorCode = res.Code()
		pkg.ErrorText = res.DeveloperMessage()
		event = entity.PackageEventTypeError
		eventDesc = pkg.Name
		pkg.Status = entity.PackageStatusFailed
	}

	// Создание транзакции.
	tx, err := p.packageRepo.BeginTX(ctx)
	if err != nil {
		return aerror.New(ctx, entity.OpenTXErrorID, err, "Ошибка при создании транзакции: %s", err.Error())
	}

	defer func() { _ = tx.RollbackIfNotCommitted() }()

	// Создание события завершения обработки пакета.
	err = p.packageRepo.WithTX(&tx).AddNewEvent(pkg.ID, event, eventDesc)
	if err != nil {
		return aerror.New(ctx, entity.InsertEventErrorID, err, "Ошибка при вставке события завершения обработки пакета в БД: %s", err.Error())
	}

	// Обновление данных пакета.
	err = p.packageRepo.WithTX(&tx).UpdatePackage(ctx, pkg)
	if err != nil {
		return aerror.New(ctx, entity.UpdatePkgErrorID, err, "Ошибка при обновлении пакета в БД: %s", err.Error())
	}

	if err = tx.Commit(); err != nil {
		return aerror.New(ctx, entity.CommitTXErrorID, err, "Ошибка фиксации транзакции: %s", err.Error())
	}

	return res
}

// Отправка пакета в сервис package-sender.
func (p *packageCreatorCoreImpl) sendPackage(ctx context.Context, pkg *entity.Package) aerror.AError {
	alogger.DebugFromCtx(ctx, fmt.Sprintf("Отправка пакета %s в сервис package-sender.", pkg.Name), nil, nil, false)

	err := p.brokerPublisher.SendPackage(ctx, pkg, p.cfg.ReceiptURL)
	if err != nil {
		return aerror.New(ctx, entity.BrokerSendErrorID, err, "Ошибка при отправке пакета в сервис package-sender.")
	}

	// Создания события подтверждающее отправку на следующий этап обработки.
	// Если не получилось - пропускаем, т.к. задача уже отправлена в package-sender.
	err = p.packageRepo.AddNewEvent(pkg.ID, entity.PackageEventTypeSent, pkg.Name)
	if err != nil {
		p.metricsCollector.AddTempError(entity.InsertEventErrorID.UserMessage())
		alogger.ErrorFromCtx(ctx, fmt.Sprintf("Ошибка вставки события в БД, подтверждающего отправку на следующий этап обработки: %s", err.Error()), err, nil, false)
	}

	return nil
}
