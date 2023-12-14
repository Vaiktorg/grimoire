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
	if a.Provider.db.Migrator().HasTable(entities.Identity{}) {
		return AlreadyExistError
	}

	return a.Provider.db.Transaction(func(tx *gorm.DB) error {
		return tx.AutoMigrate(
			&entities.Identity{},
			&entities.UserActivityLog{})
	})
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
