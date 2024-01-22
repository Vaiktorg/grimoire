package services

import (
	"context"
	"errors"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/authentity/src/models"
	"github.com/vaiktorg/grimoire/authentity/src/repo"
	"github.com/vaiktorg/grimoire/uid"
)

type ProfileService struct {
	Repo *repo.ProfileRepo
}

func NewProfileService(profileRepo *repo.ProfileRepo) ProfileService {
	return ProfileService{
		Repo: profileRepo,
	}
}

func (p *ProfileService) FetchProfile(ctx context.Context, id uid.UID) (*entities.Profile, error) {
	// Retrieve the connection pool from the context. Because the
	// r.Context().Value() method always returns an interface{} type, we
	// need to type assert it into a *sql.DB before using it.

	prof, err := p.Repo.GetProfile(ctx, id)
	if err != nil {
		return nil, errors.New("no profile found")
	}

	return prof, nil
}
func (p *ProfileService) AllProfiles(ctx context.Context) ([]*entities.Profile, error) {
	// Retrieve the connection pool from the context. Because the
	// r.Context().Value() method always returns an interface{} type, we
	// need to type assert it into a *sql.DB before using it.

	profs, err := p.Repo.Profiles(ctx)
	if err != nil {
		return nil, errors.New("no profiles found")
	}

	return profs, nil
}

func ProfileToModel(profile *entities.Profile) *models.Profile {
	var model *models.Profile
	if profile == nil {
		return nil
	}

	model = &models.Profile{
		ID:          profile.Entity.ID,
		FirstName:   *profile.FirstName,
		Initial:     *profile.Initial,
		LastName:    *profile.LastName,
		PhoneNumber: *profile.PhoneNumber,
	}

	if profile.Address != nil {
		model.Address = &models.Address{
			ID:      profile.Address.Entity.ID,
			Addr1:   *profile.Address.Addr1,
			Addr2:   *profile.Address.Addr2,
			City:    *profile.Address.City,
			State:   *profile.Address.State,
			Country: *profile.Address.Country,
			Zip:     *profile.Address.Zip,
		}
	}

	return model
}
