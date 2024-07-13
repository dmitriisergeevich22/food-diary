package packagerepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"git.astralnalog.ru/utils/postgresdb"

	"git.astralnalog.ru/edicore/package-creator/internal/entity"
	"git.astralnalog.ru/edicore/package-creator/internal/usecases/repository"
)

// Репозиторий для пакетов.
type PackageRepo interface {
	// Пинг БД.
	Ping(ctx context.Context) error
	// Начало транзакции.
	BeginTX(ctx context.Context) (postgresdb.Transaction, error)
	// Создание нового репозитория с переданной транзакцией.
	WithTX(*postgresdb.Transaction) PackageRepo
	// Создание события.
	AddNewEvent(pkgID int64, evType entity.PackageEventType, desc string) error
	// Создание пакета.
	InsertPackage(ctx context.Context, pkgType entity.PackageType, pkgName, destinationURL string, receiverIsHub bool, receiverOperatorID, senderOperatorID string) (*entity.Package, error)
	// Обновление данных пакета.
	UpdatePackage(ctx context.Context, pkg *entity.Package) error
	// Получение пакета по ID.
	SelectPackageByID(ctx context.Context, pkgID int64) (*entity.Package, error)
	// Получение пакета по имени.
	SelectPackageByName(ctx context.Context, pkgName string) (*entity.Package, error)
	// Вставка destinationURL. Возвращает id для destinationURL.
	InsertDestinationURL(ctx context.Context, url string) (int64, error)
}

type packageRepoImpl struct {
	packageEventTypeMap map[entity.PackageEventType]int64 // Таблица соотношения типов событий пакетов к их ID.
	packageStatusMap    map[entity.PackageStatus]int64    // Таблица соотношения статусов пакета к их ID.
	db                  *postgresdb.DBConnection          // Подключение к БД.
	tx                  *postgresdb.Transaction           // Текущая транзакция.
}

func NewPackageRepo(ctx context.Context, db *postgresdb.DBConnection) (PackageRepo, error) {
	packageEventTypeMap, err := getPackageEventTypeEnum(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении таблицы соотношения типов событий пакета к их ID: %s", err.Error())
	}

	packageStatusMap, err := getPackageStatusEnum(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении таблицы соотношения статусов пакета к их ID: %s", err.Error())
	}

	// Инициализация репозитория.
	return &packageRepoImpl{
		packageEventTypeMap: packageEventTypeMap,
		packageStatusMap:    packageStatusMap,
		db:                  db}, nil
}

// Получение таблицы соотношения типов события пакета к их ID.
func getPackageEventTypeEnum(_ context.Context, db *postgresdb.DBConnection) (map[entity.PackageEventType]int64, error) {
	packageEventTypeEnum := []struct {
		ID   int64  `db:"id"`
		Desc string `db:"package_event_desc"`
	}{}

	err := db.GetReadConnection().Select(&packageEventTypeEnum, "SELECT id, package_event_desc FROM package_event_enum")
	if err != nil {
		return nil, fmt.Errorf("ошибка при выборке типов событий пакета из БД: %s", err.Error())
	}

	// Таблица соотношения типа события пакета к его ID.
	packageEventTypeMap := make(map[entity.PackageEventType]int64)

	// Поиск всех типов из таблицы packageEventTypeList.
	for _, packageEventTypeDB := range packageEventTypeEnum {
		var found bool

		for _, desc := range packageEventTypeList {
			if desc == entity.PackageEventType(packageEventTypeDB.Desc) {
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("список типов событий пакета в коде не соответствует таблице package_event_enum, не найден %s", packageEventTypeDB.Desc)
		}

		// Заполнение таблицы соотношения типа события пакета к его ID.
		packageEventTypeMap[entity.PackageEventType(packageEventTypeDB.Desc)] = packageEventTypeDB.ID
	}

	// Валидация
	if len(packageEventTypeList) != len(packageEventTypeMap) {
		return nil, fmt.Errorf("список типов события пакета в коде не соответствует таблице package_event_enum")
	}

	return packageEventTypeMap, nil
}

// Получение таблицы соотношения статусов пакета к их ID.
func getPackageStatusEnum(_ context.Context, db *postgresdb.DBConnection) (map[entity.PackageStatus]int64, error) {
	packageStatusEnum := []struct {
		ID   int64  `db:"id"`
		Desc string `db:"package_status_desc"`
	}{}

	err := db.GetReadConnection().Select(&packageStatusEnum, "SELECT id, package_status_desc FROM package_status_enum")
	if err != nil {
		return nil, fmt.Errorf("ошибка при выборке статусов пакета из БД: %s", err.Error())
	}

	// Таблица соотношения статуса пакета к его ID.
	packageStatusMap := make(map[entity.PackageStatus]int64)

	// Поиск всех статусов из таблицы packageStatusList.
	for _, packageStatusDB := range packageStatusEnum {
		var found bool

		for _, desc := range packageStatusList {
			if desc == entity.PackageStatus(packageStatusDB.Desc) {
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("список статусов пакета в коде не соответствует таблице package_status_enum, не найден %s", packageStatusDB.Desc)
		}

		// Заполнение таблицы соотношения статуса пакета к его ID.
		packageStatusMap[entity.PackageStatus(packageStatusDB.Desc)] = packageStatusDB.ID
	}

	// Валидация
	if len(packageStatusList) != len(packageStatusMap) {
		return nil, fmt.Errorf("список статусов пакета в коде не соответствует таблице package_status_enum")
	}

	return packageStatusMap, nil
}

func (repo *packageRepoImpl) getReadConnection() postgresdb.QueryExecutor {
	if repo.tx != nil {
		return *repo.tx
	}

	return repo.db.GetReadConnection()
}

func (repo *packageRepoImpl) getWriteConnection() postgresdb.QueryExecutor {
	if repo.tx != nil {
		return *repo.tx
	}

	return repo.db.GetWriteConnection()
}

func (repo *packageRepoImpl) Ping(_ context.Context) error {
	return repo.getWriteConnection().Ping()
}

func (repo *packageRepoImpl) BeginTX(ctx context.Context) (postgresdb.Transaction, error) {
	return repo.db.GetWriteConnection().BeginTX(ctx)
}

func (repo *packageRepoImpl) WithTX(tx *postgresdb.Transaction) PackageRepo {
	return &packageRepoImpl{
		packageEventTypeMap: repo.packageEventTypeMap,
		packageStatusMap:    repo.packageStatusMap,
		db:                  repo.db,
		tx:                  tx}
}

func (repo *packageRepoImpl) AddNewEvent(pkgID int64, event entity.PackageEventType, desc string) error {
	query := `INSERT INTO package_events (package_id, package_event_id, description) VALUES ($1, $2, NULLIF($3, ''));`

	_, err := repo.getWriteConnection().ExecOne(query, pkgID, repo.packageEventTypeMap[event], desc)
	if err != nil {
		return err
	}

	return nil
}

func (repo *packageRepoImpl) InsertPackage(ctx context.Context, pkgType entity.PackageType, pkgName, destinationURL string,
	receiverIsHub bool, receiverOperatorID, senderOperatorID string) (*entity.Package, error) {
	pkg := entity.Package{
		Type:               pkgType,
		Name:               pkgName,
		DestinationURL:     destinationURL,
		ReceiverIsHub:      receiverIsHub,
		ReceiverOperatorID: receiverOperatorID,
		SenderOperatorID:   senderOperatorID,
		CreatedAt:          time.Now(),
		Status:             entity.PackageStatusCreated,
	}

	// Получение ID для DestinationURL.
	destinationURLID, err := repo.InsertDestinationURL(ctx, destinationURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка при добавлении destinationURL: %s", err.Error())
	}

	query := `
		INSERT INTO packages (type, name, destination_url_id, receiver_is_hub, receiver_operator_id, sender_operator_id, package_status_id, created_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	pkgIDWrapper := repository.IDWrapper{}

	err = repo.getWriteConnection().QueryAndScan(&pkgIDWrapper, query,
		pkg.Type,
		pkg.Name,
		destinationURLID,
		pkg.ReceiverIsHub,
		pkg.ReceiverOperatorID,
		pkg.SenderOperatorID,
		repo.packageStatusMap[pkg.Status],
		pkg.CreatedAt)
	if err != nil {
		return nil, err
	}

	pkg.ID = pkgIDWrapper.ID.Int64

	return &pkg, nil
}

func (repo *packageRepoImpl) UpdatePackage(ctx context.Context, pkg *entity.Package) error {
	pkgDB := convertToPackageDB(ctx, pkg, repo.packageStatusMap)

	// Обновление ошибки пакета.
	if pkgDB.ErrorCode.Valid {
		_, err := repo.getWriteConnection().ExecOne(
			`INSERT INTO package_error (package_id, error_text, error_code) VALUES ($1, $2, $3)
			ON CONFLICT (package_id) DO UPDATE SET error_text = $2, error_code = $3`,
			pkgDB.ID, pkgDB.ErrorText, pkgDB.ErrorCode)
		if err != nil {
			return fmt.Errorf("ошибка при обновлении ошибки пакета: %s", err.Error())
		}
	} else {
		_, err := repo.getWriteConnection().Exec("DELETE FROM package_error WHERE package_id = $1", pkgDB.ID)
		if err != nil {
			return fmt.Errorf("ошибка при удалении ошибки пакета: %s", err.Error())
		}
	}

	// Вставка destinationURL.
	urlID, err := repo.InsertDestinationURL(ctx, pkgDB.DestinationURL)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении destinationURL: %s", err.Error())
	}

	// Обновление данных пакета.
	query := `
		UPDATE packages SET
			type = $2,
			destination_url_id = $3,
			receiver_is_hub = $4,
			receiver_operator_id = $5,
			sender_operator_id = $6,
			package_status_id = $7,
			updated_at = current_timestamp
		WHERE id = $1`

	_, err = repo.getWriteConnection().ExecOne(query, pkgDB.ID, pkg.Type, urlID, pkgDB.ReceiverIsHub, pkgDB.ReceiverOperatorID, pkgDB.SenderOperatorID, pkgDB.StatusID)
	if err != nil {
		return err
	}

	return nil
}

const (
	queryDocflowFullSelect = `
	SELECT
		pkg.id,
		pkg.type,
		pkg.name,
		surl.destination_url,
		pkg.receiver_is_hub,
		pkg.receiver_operator_id,
		pkg.sender_operator_id,
		pkg.created_at,
		pstatus.package_status_desc,
		perr.error_text,
		perr.error_code
	FROM packages AS pkg
		LEFT JOIN package_error AS perr ON perr.package_id = pkg.id
		LEFT JOIN package_status_enum AS pstatus ON pstatus.id = pkg.package_status_id
		LEFT JOIN destination_urls AS surl ON surl.id = pkg.destination_url_id`
)

func (repo *packageRepoImpl) SelectPackageByID(ctx context.Context, pkgID int64) (*entity.Package, error) {
	var pkgDB packageDB

	query := queryDocflowFullSelect + ` WHERE pkg.id = $1`

	err := repo.getReadConnection().Get(&pkgDB, query, pkgID)
	if err != nil {
		return nil, err
	}

	return pkgDB.ConvertToEntityPackage(ctx), nil
}

func (repo *packageRepoImpl) SelectPackageByName(ctx context.Context, pkgName string) (*entity.Package, error) {
	var pkgDB packageDB

	query := queryDocflowFullSelect + ` WHERE pkg.name = $1`

	err := repo.getReadConnection().Get(&pkgDB, query, pkgName)
	if err != nil {
		return nil, err
	}

	return pkgDB.ConvertToEntityPackage(ctx), nil
}

func (repo *packageRepoImpl) InsertDestinationURL(_ context.Context, url string) (int64, error) {
	var id int64

	id, err := repo.getDestinationURLID(url)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	if id == 0 {
		id, err = repo.insertDestinationURL(url)
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

func (repo *packageRepoImpl) getDestinationURLID(url string) (int64, error) {
	var id int64

	query := `SELECT id FROM public.destination_urls WHERE destination_url = $1;`

	err := repo.getReadConnection().Get(&id, query, url)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (repo *packageRepoImpl) insertDestinationURL(url string) (int64, error) {
	var id int64

	query := `INSERT INTO public.destination_urls (destination_url) VALUES ($1::text) RETURNING id;`

	err := repo.getWriteConnection().Get(&id, query, url)
	if err != nil {
		return 0, err
	}

	return id, nil
}
