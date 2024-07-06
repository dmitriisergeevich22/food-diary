package repo

import (
	"context"
	"database/sql"
	"time"

)

type eventDB struct {
	ID        int64            // Идентификатор события.
	Type               string    `db:"type"`

	ErrorCode sql.NullString `db:"error_code"`
	ErrorText sql.NullString `db:"error_text"`
}

func convertToPackageDB(_ context.Context, in *entity.Package, packageStatusMap map[entity.PackageStatus]int64) *packageDB {
	return &packageDB{
		ID:                 in.ID,
		Type:               string(in.Type),
		Name:               in.Name,
		DestinationURL:     in.DestinationURL,
		ReceiverIsHub:      in.ReceiverIsHub,
		ReceiverOperatorID: in.ReceiverOperatorID,
		SenderOperatorID:   in.SenderOperatorID,
		StatusID:           packageStatusMap[in.Status],
		StatusDesc:         string(in.Status),
		CreatedAt:          in.CreatedAt,
		ErrorText:          repository.StringToNullString(in.ErrorText),
		ErrorCode:          repository.StringToNullString(in.ErrorCode),
	}
}

func (pkgDB packageDB) ConvertToEntityPackage(_ context.Context) *entity.Package {
	return &entity.Package{
		ID:                 pkgDB.ID,
		Type:               entity.PackageType(pkgDB.Type),
		Name:               pkgDB.Name,
		DestinationURL:     pkgDB.DestinationURL,
		ReceiverIsHub:      pkgDB.ReceiverIsHub,
		ReceiverOperatorID: pkgDB.ReceiverOperatorID,
		SenderOperatorID:   pkgDB.SenderOperatorID,
		Status:             entity.PackageStatus(pkgDB.StatusDesc),
		CreatedAt:          pkgDB.CreatedAt,
		ErrorText:          repository.NullStringToString(pkgDB.ErrorText),
		ErrorCode:          repository.NullStringToString(pkgDB.ErrorCode),
	}
}

var packageEventTypeList = []entity.PackageEventType{
	entity.PackageEventTypeCreated,
	entity.PackageEventTypeGotAgain,
	entity.PackageEventTypeReprocess,
	entity.PackageEventTypeSuccess,
	entity.PackageEventTypeSent,
	entity.PackageEventTypeError,
}

// Список статусов пакета в БД.
var packageStatusList = []entity.PackageStatus{
	entity.PackageStatusCreated,
	entity.PackageStatusSuccess,
	entity.PackageStatusFailed,
}
