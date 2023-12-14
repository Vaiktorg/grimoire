package repo

import (
	"context"
	"github.com/google/uuid"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"gorm.io/gorm"
	"sync"
)

type IdentityRepo struct {
	mu sync.Mutex
	db *gorm.DB
}

func NewIdentityRepo(db *gorm.DB) *IdentityRepo { return &IdentityRepo{db: db} }

// FindIdentityByID returns Identity when matched with a ProfileID.
func (a *IdentityRepo) FindIdentityByID(ctx context.Context, id string) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	identity := &entities.Identity{Model: entities.Model{ID: id}}
	if err := a.db.WithContext(ctx).Take(&identity).Error; err != nil {
		return nil, err
	}

	return identity, nil
}

// FindIdentityByProfileID returns Identity when matched with a ProfileID.
func (a *IdentityRepo) FindIdentityByProfileID(ctx context.Context, profileId string) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	uid, err := uuid.Parse(profileId)
	if err != nil {
		return nil, err
	}

	identity := &entities.Identity{ProfileID: uid}
	if err = a.db.WithContext(ctx).Take(&identity).Error; err != nil {
		return nil, err
	}

	return identity, nil
}

// FindIdentityByAccountID returns Identity when matched with a AccountID.
func (a *IdentityRepo) FindIdentityByAccountID(ctx context.Context, accId string) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	identity := &entities.Identity{Account: &entities.Account{Model: entities.Model{ID: accId}}}
	if err := a.db.WithContext(ctx).
		Joins("Account", a.db.Where(identity)).
		Take(&identity).Error; err != nil {
		return nil, err
	}

	return identity, nil
}
func (a *IdentityRepo) FindIdentityByAccountUsername(ctx context.Context, username string) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	identity := &entities.Identity{}
	if err := a.db.WithContext(ctx).Joins("Account", a.db.Where(&entities.Account{Username: username})).Take(&identity).Error; err != nil {
		return nil, err
	}

	return identity, nil
}
func (a *IdentityRepo) FindIdentityByAccountEmail(ctx context.Context, email string) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	identity := &entities.Identity{Account: &entities.Account{Email: email}}
	if err := a.db.WithContext(ctx).Joins("Account", a.db.Where(&entities.Account{Email: email})).Take(&identity).Error; err != nil {
		return nil, err
	}

	return identity, nil
}

func (a *IdentityRepo) AllIdentities(ctx context.Context) ([]*entities.Identity, error) {

	var bks []*entities.Identity
	if err := a.db.WithContext(ctx).Find(&bks).Error; err != nil {
		return nil, err
	}

	return bks, nil
}

func (a *IdentityRepo) Persist(ctx context.Context, identity *entities.Identity) error {
	return a.db.WithContext(ctx).Save(identity).Error
}
