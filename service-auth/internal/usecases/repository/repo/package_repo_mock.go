package packagerepo

import (
	"context"
	"crypto/rand"
	"database/sql"
	"math/big"
	"sync"
	"time"

	aerror "git.astralnalog.ru/utils/aerror"
	"git.astralnalog.ru/utils/postgresdb"

	"git.astralnalog.ru/edicore/package-creator/internal/entity"
)

// Имитация репозитория пакетов. Подключение и контекст может быть nil.
func NewPackageRepoMock(_ context.Context, _ *postgresdb.DBConnection) (PackageRepo, aerror.AError) {
	return packageRepoMockImpl{
		packages:        make(map[int64]*entity.Package),
		packageEvents:   make(map[int64][]entity.PackageEvent),
		destinationURLs: make(map[string]int64),
		m:               &sync.Mutex{}}, nil
}

// Имплементация имитации репозитория.
type packageRepoMockImpl struct {
	packages        map[int64]*entity.Package       // Таблица пакетов
	packageEvents   map[int64][]entity.PackageEvent // Таблица событий
	destinationURLs map[string]int64                // Таблица DestinationURL (key: url; value: id)
	m               *sync.Mutex
}

func (repo packageRepoMockImpl) Ping(_ context.Context) error {
	repo.m.Lock()
	defer repo.m.Unlock()

	return nil
}

func (repo packageRepoMockImpl) BeginTX(_ context.Context) (postgresdb.Transaction, error) {
	return postgresdb.TransactionMock{}, nil
}

func (repo packageRepoMockImpl) WithTX(_ *postgresdb.Transaction) PackageRepo {
	return repo
}

func (repo packageRepoMockImpl) AddNewEvent(pkgID int64, evType entity.PackageEventType, desc string) error {
	repo.m.Lock()
	defer repo.m.Unlock()

	_, ok := repo.packageEvents[pkgID]
	if !ok {
		repo.packageEvents[pkgID] = []entity.PackageEvent{}
	}

	repo.packageEvents[pkgID] = append(repo.packageEvents[pkgID], entity.PackageEvent{
		PackageID:        pkgID,
		CreatedAt:        time.Now(),
		PackageEventType: evType,
		Description:      desc,
	})

	return nil
}

func (repo packageRepoMockImpl) InsertPackage(_ context.Context, pkgType entity.PackageType, pkgName, destinationURL string,
	 receiverIsHub bool, receiverOperatorID, senderOperatorID string) (*entity.Package, error) {
	repo.m.Lock()
	defer repo.m.Unlock()

	maxID := 9999
	pkgID, _ := rand.Int(rand.Reader, big.NewInt(int64(maxID)))

	pkg := entity.Package{
		ID:             pkgID.Int64(),
		Type:           pkgType,
		Name:           pkgName,
		DestinationURL: destinationURL,
		ReceiverIsHub:  receiverIsHub,
		ReceiverOperatorID: receiverOperatorID,
		SenderOperatorID:   senderOperatorID,
		CreatedAt:      time.Now(),
		Status:         entity.PackageStatusCreated,
	}

	repo.packages[pkgID.Int64()] = &pkg

	return &pkg, nil
}

func (repo packageRepoMockImpl) InsertDestinationURL(_ context.Context, destinationURL string) (int64, error) {
	repo.m.Lock()
	defer repo.m.Unlock()

	id, ok := repo.destinationURLs[destinationURL]
	if ok {
		return id, nil
	}

	maxID := 9999
	newID, _ := rand.Int(rand.Reader, big.NewInt(int64(maxID)))
	repo.destinationURLs[destinationURL] = newID.Int64()

	return newID.Int64(), nil
}

func (repo packageRepoMockImpl) UpdatePackage(_ context.Context, pkg *entity.Package) error {
	repo.m.Lock()
	defer repo.m.Unlock()

	repo.packages[pkg.ID] = pkg

	return nil
}

func (repo packageRepoMockImpl) SelectPackageByID(_ context.Context, id int64) (*entity.Package, error) {
	repo.m.Lock()
	defer repo.m.Unlock()

	res, ok := repo.packages[id]
	if !ok {
		return nil, sql.ErrNoRows
	}

	return res, nil
}

func (repo packageRepoMockImpl) SelectPackageByName(_ context.Context, name string) (*entity.Package, error) {
	for _, pkg := range repo.packages {
		if pkg.Name == name {
			return pkg, nil
		}
	}

	return nil, sql.ErrNoRows
}
