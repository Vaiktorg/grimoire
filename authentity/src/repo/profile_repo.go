package repo

import (
	"context"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/uid"
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

func (a *ProfileRepo) GetProfile(ctx context.Context, ID uid.UID) (*entities.Profile, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	prof := &entities.Profile{}

	if err := a.db.WithContext(ctx).Find(&prof, "id = ?", ID).Error; err != nil {
		return nil, err
	}

	return prof, nil
}

func (a *ProfileRepo) Profiles(ctx context.Context) ([]*entities.Profile, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var bks []*entities.Profile
	if err := a.db.WithContext(ctx).Find(&bks).Error; err != nil {
		return nil, err
	}

	return bks, nil
}
