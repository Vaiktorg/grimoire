package src

import (
	"github.com/vaiktorg/grimoire/authentity/src/repo"
	"github.com/vaiktorg/grimoire/authentity/src/services"
	"gorm.io/gorm"
	"sync"
)

type DataProvider struct {
	Mutex            sync.Mutex
	migrator         gorm.Migrator
	ProfileService   services.ProfileService
	IdentityService  services.IdentityService
	AccountsService  services.AccountService
	ResourcesService services.ResourceService
}

func NewDataProvider(db *gorm.DB) *DataProvider {
	return &DataProvider{
		migrator:         db.Migrator(),
		AccountsService:  services.NewAccountService(repo.NewAccountRepo(db)),
		IdentityService:  services.NewIdentityService(repo.NewIdentityRepo(db)),
		ProfileService:   services.NewProfileService(repo.NewProfileRepo(db)),
		ResourcesService: services.NewResourceService(repo.NewIdentityRepo(db)),
	}
}
