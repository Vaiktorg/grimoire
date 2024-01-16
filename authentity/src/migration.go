package src

import (
	"errors"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"gorm.io/gorm"
)

var (
	AlreadyExistError = errors.New("tables already in database")
)

func (a *Authentity) Migrate() error {
	if a.provider.migrator.HasTable(entities.Identity{}) {
		return AlreadyExistError
	}

	return a.provider.migrator.AutoMigrate(
		&entities.Identity{
			Entity:  entities.Entity{},
			Profile: &entities.Profile{},
			Account: &entities.Account{},
		},
		&entities.UserActivityLog{},
	)
}

func (a *Authentity) Drop(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(
			&entities.Account{},
			&entities.Profile{},
			&entities.Identity{},
			&entities.Address{},
		)
	})
}
