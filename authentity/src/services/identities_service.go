package services

import (
	"context"
	"encoding/json"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/authentity/src/repo"
	"net/http"
)

// IdentitiesHandler Return identities
func IdentitiesHandler(service *IdentityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass the request context onto the database layer.
		bks, err := service.FetchIdentities(r.Context())
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

// IdentityHandler ...
func IdentityHandler(service *IdentityService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass the request context onto the database layer.
		id := r.URL.Query().Get("id")

		if id == "" {
			http.Error(w, "id not provided", http.StatusInternalServerError)
			return
		}

		bks, err := service.FetchIdentity(r.Context(), id)
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

// ====================================================================================================

type IdentityService struct {
	Repo *repo.IdentityRepo
}

func NewIdentityService(identityRepo *repo.IdentityRepo) IdentityService {
	return IdentityService{
		Repo: identityRepo,
	}
}

func (i *IdentityService) FetchIdentities(ctx context.Context) ([]*entities.Identity, error) {
	// To views

	return i.Repo.AllIdentities(ctx)
}
func (i *IdentityService) FetchIdentity(ctx context.Context, ID string) (*entities.Identity, error) {
	// To views

	return i.Repo.FindIdentityByID(ctx, ID)
}
func (i *IdentityService) FetchIdentityByAccountID(ctx context.Context, ID string) (*entities.Identity, error) {
	// To views

	return i.Repo.FindIdentityByAccountID(ctx, ID)
}
func (i *IdentityService) FetchIdentityByEmail(ctx context.Context, email string) (*entities.Identity, error) {
	// To views

	return i.Repo.FindIdentityByAccountEmail(ctx, email)
}
func (i *IdentityService) FetchIdentityByUsername(ctx context.Context, username string) (*entities.Identity, error) {
	// To views

	return i.Repo.FindIdentityByAccountUsername(ctx, username)
}
func (i *IdentityService) Persist(ctx context.Context, identity *entities.Identity) error {
	return i.Repo.Persist(ctx, identity)
}
