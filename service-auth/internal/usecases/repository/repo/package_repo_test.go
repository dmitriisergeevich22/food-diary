package packagerepo_test

import (
	"context"
	"testing"

	"git.astralnalog.ru/utils/postgresdb"

	"git.astralnalog.ru/edicore/package-creator/internal/entity"
	"git.astralnalog.ru/edicore/package-creator/internal/usecases/repository/packagerepo"
)

// Поднимает postgres в контейнере и накатывает на него миграции через goose. должен быть запущен docker и установлен goose.
func TestPackageRepo(t *testing.T) {
	ctx := context.Background()
	// Подключение к БД в postgres.
	conn, _, err := postgresdb.GetMockConn("../../../../migrations/package_creator_db")
	if err != nil {
		t.Fatalf("ошибка создания соединения. обычно не включен докер, или неправильно указан путь к миграциям, или не установлен goose. %s", err.Error())
	}

	packageRepo, err := packagerepo.NewPackageRepo(ctx, conn)
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = packageRepo.Ping(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Проверка добавления события пакета.
	testAddNewEvent(conn, packageRepo, t)

	// Проверка добавления пакета.
	testInsertPackage(conn, packageRepo, t)

	// Проверка обновления статуса пакета.
	testUpdatePackageStatus(conn, packageRepo, t)

	// Проверка получения пакета по ID.
	testSelectPackageByID(conn, packageRepo, t)

	// Проверка получения пакета по имени.
	testSelectPackageByName(conn, packageRepo, t)
}

func testAddNewEvent(conn *postgresdb.DBConnection, repo packagerepo.PackageRepo, t *testing.T) {
	ctx := context.Background()
	// Вставка тестового пакета, без него нельзя будет создать событие пакета.
	pkg, err := repo.InsertPackage(ctx, entity.PackageTypeMessages, "testAddNewEvent", "destinationURL", false, "Code1", "Code2")
	if err != nil {
		t.Fatalf("testAddNewEvent ошибка вставки тестового пакета: %s", err.Error())
	}

	// Вставка события создания пакета.
	err = repo.AddNewEvent(pkg.ID, entity.PackageEventTypeCreated, pkg.Name)
	if err != nil {
		t.Fatalf("testAddNewEvent ошибка создания события: %s", err.Error())
	}

	// Проверка существования события создания пакета. Важно что бы событие было одно.
	_, err = conn.GetReadConnection().ExecOne("SELECT id FROM package_events WHERE package_id = $1", pkg.ID)
	if err != nil {
		t.Fatalf("testAddNewEvent ошибка получения события: %s", err.Error())
	}

	// Вставка ошибочного события.
	err = repo.AddNewEvent(0, entity.PackageEventTypeCreated, "")
	if err == nil {
		t.Fatalf("testAddNewEvent успешная вставка ошибочного события.")
	}
}

func testInsertPackage(conn *postgresdb.DBConnection, repo packagerepo.PackageRepo, t *testing.T) {
	ctx := context.Background()

	// Вставка пакета.
	pkg, err := repo.InsertPackage(ctx, entity.PackageTypeMessages, "testInsertPackageName", "destinationURL", false, "Code1", "Code2")
	if err != nil {
		t.Fatalf("testInsertPackage ошибка вставки пакета: %s", err.Error())
	}

	// Вставка дубликата.
	_, err = repo.InsertPackage(ctx, entity.PackageTypeMessages, "testInsertPackageName", "destinationURL", true, "Code1", "Code2")
	if err == nil {
		t.Fatalf("testInsertPackage вставлен дубликат пакета.")
	}

	if pkg == nil {
		t.Fatalf("testInsertPackage возвращен пакет nil")
	}

	// Проверка существования пакета.
	var rows []int64

	err = conn.GetReadConnection().Select(&rows, "SELECT id FROM packages WHERE name = $1", pkg.Name)
	if err != nil {
		t.Fatalf("AddNewEvent ошибка получения пакета testInsertPackageName: %s", err.Error())
	}

	if len(rows) != 1 {
		t.Fatalf("AddNewEvent неверное количество выбранных пакетов по имени testInsertPackageName: %d > 1", len(rows))
	}

	if rows[0] != pkg.ID {
		t.Fatalf("AddNewEvent ошибка получения пакета по имени: %d != %d", rows[0], pkg.ID)
	}
}

func testUpdatePackageStatus(_ *postgresdb.DBConnection, repo packagerepo.PackageRepo, t *testing.T) {
	ctx := context.Background()

	// Вставка пакета.
	pkg, err := repo.InsertPackage(ctx, entity.PackageTypeMessages, "testUpdatePackageStatus", "destinationURL", false, "Code1", "Code2")
	if err != nil {
		t.Fatalf("testUpdatePackageStatus ошибка вставки пакета: %s", err.Error())
	}

	if pkg == nil {
		t.Fatalf("testUpdatePackageStatus возвращен пакет nil")
	}

	// Обновление пакета.
	pkg.DestinationURL = "testNewDestinationURL"
	pkg.Type = entity.PackageTypeInvitation
	pkg.Status = entity.PackageStatusSuccess
	pkg.ErrorCode = "testErrorCode"
	pkg.ErrorText = "testErrorText"

	err = repo.UpdatePackage(ctx, pkg)
	if err != nil {
		t.Fatalf("testUpdatePackageStatus ошибка обновления пакета: %s", err.Error())
	}

	// Проверка итога.
	getPkg, err := repo.SelectPackageByID(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("testUpdatePackageStatus ошибка получения пакета: %s", err.Error())
	}

	if getPkg == nil {
		t.Fatalf("testUpdatePackageStatus получен пустой пакет")
	}

	if getPkg.Type != pkg.Type {
		t.Fatalf("testUpdatePackageStatus ошибка обновления типа пакета: %s != %s", getPkg.Type, pkg.Type)
	}

	if getPkg.DestinationURL != pkg.DestinationURL {
		t.Fatalf("testUpdatePackageStatus ошибка обновления destinationURL пакета: %s != %s", getPkg.DestinationURL, pkg.DestinationURL)
	}

	if getPkg.Status != pkg.Status {
		t.Fatalf("testUpdatePackageStatus ошибка обновления статуса пакета: %s != %s", getPkg.Status, pkg.Status)
	}

	if getPkg.ErrorCode != pkg.ErrorCode {
		t.Fatalf("testUpdatePackageStatus ошибка обновления кода ошибки пакета: %s != %s", getPkg.ErrorCode, pkg.ErrorCode)
	}

	if getPkg.ErrorText != pkg.ErrorText {
		t.Fatalf("testUpdatePackageStatus ошибка обновления текста ошибки пакета: %s != %s", getPkg.ErrorText, pkg.ErrorText)
	}
}

func testSelectPackageByID(_ *postgresdb.DBConnection, repo packagerepo.PackageRepo, t *testing.T) {
	ctx := context.Background()

	// Вставка пакета.
	pkg, err := repo.InsertPackage(ctx, entity.PackageTypeMessages, "testSelectPackageByID", "destinationURL", false, "Code1", "Code2")
	if err != nil {
		t.Fatalf("testSelectPackageByID ошибка вставки пакета: %s", err.Error())
	}

	if pkg == nil {
		t.Fatalf("testSelectPackageByID возвращен пакет nil")
	}

	// Получение пакета по ID.
	getPkg, err := repo.SelectPackageByID(ctx, pkg.ID)
	if err != nil {
		t.Fatalf("testSelectPackageByID ошибка получения пакета: %s", err.Error())
	}

	// Проверка итога.
	if getPkg == nil {
		t.Fatalf("testSelectPackageByID получен пустой пакет")
	}

	if getPkg.ID != pkg.ID {
		t.Fatalf("testSelectPackageByID несовпадение ID пакета: %d != %d", getPkg.ID, pkg.ID)
	}

	if getPkg.Type != pkg.Type {
		t.Fatalf("testSelectPackageByID несовпадение типа пакета: %s != %s", getPkg.Type, pkg.Type)
	}

	if getPkg.Name != pkg.Name {
		t.Fatalf("testSelectPackageByID несовпадение имени пакета: %s != %s", getPkg.Name, pkg.Name)
	}

	if getPkg.Status != pkg.Status {
		t.Fatalf("testSelectPackageByID несовпадение статуса пакета: %s != %s", getPkg.Status, pkg.Status)
	}

	if getPkg.ErrorCode != pkg.ErrorCode {
		t.Fatalf("testSelectPackageByID несовпадение кода ошибки пакета: %s != %s", getPkg.ErrorCode, pkg.ErrorCode)
	}

	if getPkg.ErrorText != pkg.ErrorText {
		t.Fatalf("testSelectPackageByID несовпадение текста ошибки пакета: %s != %s", getPkg.ErrorText, pkg.ErrorText)
	}
}

func testSelectPackageByName(_ *postgresdb.DBConnection, repo packagerepo.PackageRepo, t *testing.T) {
	ctx := context.Background()

	// Вставка пакета.
	pkg, err := repo.InsertPackage(ctx, entity.PackageTypeMessages, "testSelectPackageByName", "destinationURL", false, "Code1", "Code2")
	if err != nil {
		t.Fatalf("testSelectPackageByName ошибка вставки пакета: %s", err.Error())
	}

	if pkg == nil {
		t.Fatalf("testSelectPackageByName возвращен пакет nil")
	}

	// Получение пакета по ID.
	getPkg, err := repo.SelectPackageByName(ctx, pkg.Name)
	if err != nil {
		t.Fatalf("testSelectPackageByName ошибка получения пакета: %s", err.Error())
	}

	// Проверка итога.
	if getPkg == nil {
		t.Fatalf("testSelectPackageByName получен пустой пакет")
	}

	if getPkg.ID != pkg.ID {
		t.Fatalf("testSelectPackageByName несовпадение ID пакета: %d != %d", getPkg.ID, pkg.ID)
	}

	if getPkg.Type != pkg.Type {
		t.Fatalf("testSelectPackageByName несовпадение типа пакета: %s != %s", getPkg.Type, pkg.Type)
	}

	if getPkg.Name != pkg.Name {
		t.Fatalf("testSelectPackageByName несовпадение имени пакета: %s != %s", getPkg.Name, pkg.Name)
	}

	if getPkg.Status != pkg.Status {
		t.Fatalf("testSelectPackageByName несовпадение статуса пакета: %s != %s", getPkg.Status, pkg.Status)
	}

	if getPkg.ErrorCode != pkg.ErrorCode {
		t.Fatalf("testSelectPackageByName несовпадение кода ошибки пакета: %s != %s", getPkg.ErrorCode, pkg.ErrorCode)
	}

	if getPkg.ErrorText != pkg.ErrorText {
		t.Fatalf("testSelectPackageByName несовпадение текста ошибки пакета: %s != %s", getPkg.ErrorText, pkg.ErrorText)
	}
}
