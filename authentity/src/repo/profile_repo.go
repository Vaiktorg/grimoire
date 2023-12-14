package repo

import (
	"context"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"gorm.io/gorm"
	"sync"
)

type ProfileRepo struct {
	mu sync.Mutex
	db *gorm.DB
}

func NewProfileRepo(db *gorm.DB) *ProfileRepo {
	return &ProfileRepo{db: db}
}

func (a *ProfileRepo) GetProfile(ctx context.Context, id string) *entities.Profile {
	a.mu.Lock()
	defer a.mu.Unlock()

	prof := &entities.Profile{
		Model: entities.Model{ID: id},
	}

	a.db.WithContext(ctx).Find(prof)

	return prof
}

func (a *ProfileRepo) Profiles(ctx context.Context) []*entities.Profile {
	a.mu.Lock()
	defer a.mu.Unlock()

	var bks []*entities.Profile
	a.db.WithContext(ctx).Find(&bks)

	return bks
}
