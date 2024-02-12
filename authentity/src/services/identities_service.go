package services

import (
	"context"
	"errors"
	"github.com/vaiktorg/grimoire/authentity/src/entities"
	"github.com/vaiktorg/grimoire/authentity/src/models"
	"github.com/vaiktorg/grimoire/authentity/src/repo"
	"github.com/vaiktorg/grimoire/gwt"
)

type IdentityService struct {
	Repo *repo.IdentityRepo
}

func NewIdentityService(identityRepo *repo.IdentityRepo) IdentityService {
	return IdentityService{
		Repo: identityRepo,
	}
}

func (i *IdentityService) FetchIdentities(ctx context.Context) ([]models.Identity, error) {
	// To views
	identities, err := i.Repo.AllIdentities(ctx)
	if err != nil {
		return nil, err
	}

	var retIdentities = make([]models.Identity, len(identities))
	var identity *models.Identity

	for _, entity := range identities {
		identity, err = IdentityToModel(entity)
		if err != nil {
			return nil, err
		}

		retIdentities = append(retIdentities, *identity)
	}

	return retIdentities, nil
}
func (i *IdentityService) FetchIdentity(ctx context.Context, ID string) (*models.Identity, error) {
	// To views
	identity, err := i.Repo.FindIdentityByID(ctx, ID)
	if err != nil {
		return nil, err
	}

	return IdentityToModel(identity)
}
func (i *IdentityService) FetchIdentityByAccountID(ctx context.Context, ID string) (*models.Identity, error) {
	// To views
	identity, err := i.Repo.FindIdentityByAccountID(ctx, ID)
	if err != nil {
		return nil, err
	}

	if identity.Account == nil {
		return nil, errors.New("no identity found for account id " + ID)
	}

	return IdentityToModel(identity)
}
func (i *IdentityService) FetchIdentityByEmail(ctx context.Context, email string) (*models.Identity, error) {
	// To views

	identity, err := i.Repo.FindIdentityByAccountEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return IdentityToModel(identity)
}
func (i *IdentityService) FetchIdentityByUsername(ctx context.Context, username string) (*models.Identity, error) {
	// To views

	identity, err := i.Repo.FindIdentityByAccountUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	return IdentityToModel(identity)
}
func (i *IdentityService) Persist(ctx context.Context, identity *models.Identity) error {
	return i.Repo.Persist(ctx, IdentityToEntity(identity))
}
func (i *IdentityService) Updates(ctx context.Context, identity *models.Identity) error {
	return i.Repo.Update(ctx, IdentityToEntity(identity))
}

// ====================================================================================================

func IdentityToModel(identity *entities.Identity) (*models.Identity, error) {
	resource := new(gwt.Resources)
	if err := resource.Deserialize([]byte(*identity.Resources)); err != nil {
		return nil, err
	}

	model := &models.Identity{
		ID:        identity.Entity.ID,
		Resources: resource,
	}

	model.Profile = ProfileToModel(identity.Profile)
	model.Account = AccountToModel(identity.Account)

	return model, nil
}
func IdentityToEntity(identity *models.Identity) *entities.Identity {
	res := string(identity.Resources.Serialize())
	entity := &entities.Identity{
		Entity:    entities.Entity{ID: identity.ID},
		Resources: &res,
	}

	if identity.Profile != nil {
		entity.ProfileID = identity.Profile.ID
		entity.Profile = &entities.Profile{
			Entity:      entities.Entity{ID: identity.Profile.ID},
			FirstName:   &identity.Profile.FirstName,
			Initial:     &identity.Profile.Initial,
			LastName:    &identity.Profile.LastName,
			PhoneNumber: &identity.Profile.PhoneNumber,
		}

		if identity.Profile.Address != nil {
			entity.Profile.AddressID = identity.Profile.Address.ID
			entity.Profile.Address = &entities.Address{
				Entity:  entities.Entity{ID: identity.Profile.Address.ID},
				Addr1:   &identity.Profile.Address.Addr1,
				Addr2:   &identity.Profile.Address.Addr2,
				City:    &identity.Profile.Address.City,
				State:   &identity.Profile.Address.State,
				Country: &identity.Profile.Address.Country,
				Zip:     &identity.Profile.Address.Zip,
			}
		}
	}

	if identity.Account != nil {
		entity.AccountID = identity.Account.ID
		entity.Account = &entities.Account{
			Entity:    entities.Entity{ID: identity.Account.ID},
			Username:  identity.Account.Username,
			Email:     &identity.Account.Email,
			Signature: &identity.Account.Signature,
			Password:  &identity.Account.Password,
		}
	}

	return entity
}
