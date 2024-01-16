package services

import (
	"context"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/authentity/src/models"
	"github.com/vaiktorg/grimoire/authentity/src/repo"
)

type AccountService struct {
	Repo *repo.AccountRepo
}

func NewAccountService(accountRepo *repo.AccountRepo) AccountService {
	return AccountService{Repo: accountRepo}
}

func (a *AccountService) GetAccount(ctx context.Context, username, email string) (*models.Account, error) {
	acc, err := a.Repo.FindAccount(ctx, username, email)
	if err != nil {
		return nil, err
	}

	return AccountToModel(acc), nil
}
func (a *AccountService) FindAccountByUsername(ctx context.Context, username string) (*models.Account, error) {
	acc, err := a.Repo.FindAccountByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	return AccountToModel(acc), nil
}
func (a *AccountService) AccountHasUsername(ctx context.Context, username string) error {
	return a.Repo.AccountHasUsername(ctx, username)
}
func (a *AccountService) FindAccountByEmail(ctx context.Context, email string) (*models.Account, error) {
	acc, err := a.Repo.FindAccountByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return AccountToModel(acc), nil
}
func (a *AccountService) AllAccounts(ctx context.Context) ([]models.Account, error) {
	accounts, err := a.Repo.AllAccounts(ctx)
	if err != nil {
		return nil, err
	}

	accountsRet := make([]models.Account, len(accounts))
	for _, acc := range accounts {
		accountsRet = append(accountsRet, *AccountToModel(acc))
	}

	return accountsRet, err
}
func (a *AccountService) Updates(ctx context.Context, account *models.Account) error {
	return a.Repo.Update(ctx, AccountToEntity(account))
}

func AccountToModel(account *entities.Account) *models.Account {
	if account == nil {
		return nil
	}

	return &models.Account{
		ID:        account.Entity.ID,
		Username:  account.Username,
		Email:     *account.Email,
		Signature: *account.Signature,
		Password:  *account.Password,
	}
}
func AccountToEntity(account *models.Account) *entities.Account {
	if account == nil {
		return nil
	}

	return &entities.Account{
		Entity: entities.Entity{
			ID: account.ID,
		},
		Username:  account.Username,
		Email:     &account.Email,
		Signature: &account.Signature,
		Password:  &account.Password,
	}
}
