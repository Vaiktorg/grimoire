package repo

import (
	"context"
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
func (a *AccountRepo) FindAccountByUsername(ctx context.Context, username string) *entities.Account {
	a.mu.Lock()
	defer a.mu.Unlock()

	account := &entities.Account{Username: username}
	a.db.WithContext(ctx).Take(&account, "username = ?", username)

	return account
}

// FindAccountByEmail returns Account when matched with an email.
func (a *AccountRepo) FindAccountByEmail(ctx context.Context, email string) *entities.Account {
	a.mu.Lock()
	defer a.mu.Unlock()

	account := &entities.Account{}
	a.db.WithContext(ctx).Take(&account, "email = ?", email)

	return account
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
	a.db.WithContext(ctx).Find(&accounts)

	return accounts, nil
}

func (a *AccountRepo) SaveAccount(ctx context.Context, account *entities.Account) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.db.WithContext(ctx).Save(account)

}
