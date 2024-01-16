package repo

import (
	"context"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/uid"
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

	identity := &entities.Identity{}
	if err := a.db.WithContext(ctx).Take(&identity, "id  = ?", id).Error; err != nil {
		return nil, err
	}

	return identity, nil
}

// FindIdentityByProfileID returns Identity when matched with a ProfileID.
func (a *IdentityRepo) FindIdentityByProfileID(ctx context.Context, profileId uid.UID) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	identity := &entities.Identity{}
	if err := a.db.WithContext(ctx).Find(&identity, "profile_id = ?", profileId).Error; err != nil {
		return nil, err
	}

	return identity, nil
}

// FindIdentityByAccountID returns Identity when matched with a AccountID.
func (a *IdentityRepo) FindIdentityByAccountID(ctx context.Context, accId string) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	identity := &entities.Identity{}
	if err := a.db.WithContext(ctx).Joins("Account", a.db.Where("account_id = ?", accId)).Find(&identity).Error; err != nil {
		return nil, err
	}

	return identity, nil
}
func (a *IdentityRepo) FindIdentityByAccountUsername(ctx context.Context, username string) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	identity := &entities.Identity{}
	if err := a.db.WithContext(ctx).Joins(
		"Account",
		a.db.Where("username = ?", username)).Take(&identity).Error; err != nil {
		return nil, err
	}

	return identity, nil
}
func (a *IdentityRepo) FindIdentityByAccountEmail(ctx context.Context, email string) (*entities.Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	identity := &entities.Identity{}
	if err := a.db.WithContext(ctx).Joins("Account", a.db.Where("email = ?", email)).Take(&identity).Error; err != nil {
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
func (a *IdentityRepo) Update(ctx context.Context, identity *entities.Identity) error {
	return a.db.WithContext(ctx).Model(&identity).Select("Account", "Profile", "Resources").Updates(identity).Error
}
