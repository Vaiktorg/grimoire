package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/authentity/src/repo"
	"net/http"
)

type ProfileService struct {
	Repo *repo.ProfileRepo
}

func NewProfileService(profileRepo *repo.ProfileRepo) ProfileService {
	return ProfileService{
		Repo: profileRepo,
	}
}

// ProfilesHandler Return profiles
func ProfilesHandler(service *ProfileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass the request context onto the database layer.
		bks, err := service.AllProfiles(r.Context())
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

func ProfileHandler(service *ProfileService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Pass the request context onto the database layer.
		id := r.URL.Query().Get("id")

		if id == "" {
			http.Error(w, "id not provided", http.StatusInternalServerError)
			return
		}

		bks, err := service.FetchProfile(r.Context(), id)
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

//==================================

func (p *ProfileService) FetchProfile(ctx context.Context, id string) (*entities.Profile, error) {
	// Retrieve the connection pool from the context. Because the
	// r.Context().Value() method always returns an interface{} type, we
	// need to type assert it into a *sql.DB before using it.

	prof := p.Repo.GetProfile(ctx, id)
	if prof == nil {
		return nil, errors.New("no profile found")
	}

	return prof, nil
}
func (p *ProfileService) AllProfiles(ctx context.Context) ([]*entities.Profile, error) {
	// Retrieve the connection pool from the context. Because the
	// r.Context().Value() method always returns an interface{} type, we
	// need to type assert it into a *sql.DB before using it.

	profs := p.Repo.Profiles(ctx)
	if profs == nil {
		return nil, errors.New("no profiles found")
	}

	return profs, nil
}
