package src

import (
	"github.com/vaiktorg/grimoire/authentity/src/repo"
	"github.com/vaiktorg/grimoire/authentity/src/services"
	"gorm.io/gorm"
	"sync"
)

type DataProvider struct {
	Mutex           sync.Mutex
	db              *gorm.DB
	ProfileService  services.ProfileService
	IdentityService services.IdentityService
	AccountsService services.AccountService
}

func NewDataProvider(db *gorm.DB) *DataProvider {
	return &DataProvider{
		db:              db,
		AccountsService: services.NewAccountService(repo.NewAccountRepo(db)),
		IdentityService: services.NewIdentityService(repo.NewIdentityRepo(db)),
		ProfileService:  services.NewProfileService(repo.NewProfileRepo(db)),
	}
}
