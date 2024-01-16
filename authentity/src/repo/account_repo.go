package repo

import (
	"context"
	"errors"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"sync"

	"gorm.io/gorm"
)

type AccountRepo struct {
	mu sync.Mutex
	db *gorm.DB
}

func NewAccountRepo(db *gorm.DB) *AccountRepo {
	return &AccountRepo{
		db: db,
	}
}

// FindAccountByUsername returns Account when matched with a username.
func (a *AccountRepo) FindAccountByUsername(ctx context.Context, username string) (*entities.Account, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	account := &entities.Account{}
	if err := a.db.WithContext(ctx).Take(&account, "username = ?", username).Error; err != nil {
		return nil, err
	}

	return account, nil
}

func (a *AccountRepo) AccountHasUsername(ctx context.Context, username string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var acc *entities.Account
	if err := a.db.WithContext(ctx).Find(&acc, "username = ?", username).Error; err != nil {
		return err
	}

	if acc == nil {
		return errors.New("account not found")
	}

	return nil
}

// FindAccountByEmail returns Account when matched with an email.
func (a *AccountRepo) FindAccountByEmail(ctx context.Context, email string) (*entities.Account, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	account := &entities.Account{}
	if err := a.db.WithContext(ctx).Take(&account, "email = ?", email).Error; err != nil {
		return nil, err
	}

	return account, nil
}

func (a *AccountRepo) FindAccount(ctx context.Context, uname, email string) (*entities.Account, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	account := entities.Account{}
	if err := a.db.WithContext(ctx).Take(&account, "username = ? or email = ?", uname, email).Error; err != nil {
		return nil, err
	}

	return &account, nil
}

func (a *AccountRepo) AllAccounts(ctx context.Context) ([]*entities.Account, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var accounts []*entities.Account
	if err := a.db.WithContext(ctx).Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

func (a *AccountRepo) Persist(ctx context.Context, account *entities.Account) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.db.WithContext(ctx).Save(account).Error
}
func (a *AccountRepo) Update(ctx context.Context, account *entities.Account) error {
	return a.db.WithContext(ctx).Model(&account).Select("Signature", "Profile", "Resources").Updates(account).Error
}
