package services

import (
	"context"
	"encoding/json"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/authentity/src/repo"
	"net/http"
)

type AccountService struct {
	Repo *repo.AccountRepo
}

func NewAccountService(accountRepo *repo.AccountRepo) AccountService {
	return AccountService{Repo: accountRepo}
}

// AccountsHandler Return profiles
func AccountsHandler(service *AccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bks, err := service.AllAccounts(r.Context())
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = json.NewEncoder(w).Encode(bks)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}

func AccountHandler(service *AccountService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := &struct {
			Username string `json:"username"`
			Email    string `json:"email"`
			ID       string `json:"id"`
		}{}

		err := json.NewDecoder(r.Body).Decode(info)
		if err != nil {
			http.Error(w, http.StatusText(500), 500)
			return
		}

		account, err := service.GetAccount(r.Context(), info.Username, info.Email)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = json.NewEncoder(w).Encode(account)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	}
}

//==================================

func (a *AccountService) GetAccount(ctx context.Context, username, email string) (*entities.Account, error) {
	return a.Repo.FindAccount(ctx, username, email)
}
func (a *AccountService) AllAccounts(ctx context.Context) ([]*entities.Account, error) {
	return a.Repo.AllAccounts(ctx)
}
