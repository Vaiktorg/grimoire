package services

import (
	"context"
	"errors"
	"github.com/vaiktorg/grimoire/authentity/src/repo"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/uid"
)

type ResourceService struct {
	Repo *repo.IdentityRepo
}

func NewResourceService(identityRepo *repo.IdentityRepo) ResourceService {
	return ResourceService{Repo: identityRepo}
}

func (r *ResourceService) ResourcesByIdentityID(ctx context.Context, id uid.UID) (*gwt.Resources, error) {
	identity, err := r.Repo.FindIdentityByID(ctx, string(id))
	if err != nil {
		return nil, err
	}

	resources := &gwt.Resources{}
	err = resources.Deserialize([]byte(*identity.Resources))
	if err != nil {
		return nil, err
	}

	return resources, nil
}
func (r *ResourceService) ResourcesByAccountEmail(ctx context.Context, email uid.UID) (*gwt.Resources, error) {
	identity, err := r.Repo.FindIdentityByAccountEmail(ctx, string(email))
	if err != nil {
		return nil, err
	}

	resources := &gwt.Resources{}
	err = resources.Deserialize([]byte(*identity.Resources))
	if err != nil {
		return nil, err
	}

	return resources, nil
}
func (r *ResourceService) ResourcesByAccountUsername(ctx context.Context, username uid.UID) (*gwt.Resources, error) {
	identity, err := r.Repo.FindIdentityByAccountUsername(ctx, string(username))
	if err != nil {
		return nil, err
	}

	resources := &gwt.Resources{}
	err = resources.Deserialize([]byte(*identity.Resources))
	if err != nil {
		return nil, err
	}

	return resources, nil
}

// GetResourceByID fetches a specific resource by its ID
func (r *ResourceService) GetResourceByID(ctx context.Context, userID uid.UID, resourceID uid.UID) (*gwt.Resource, error) {
	// Fetch resources and filter by resourceID
	identity, err := r.Repo.FindIdentityByID(ctx, string(userID))
	if err != nil {
		return nil, err
	}

	resources := &gwt.Resources{}
	err = resources.Deserialize([]byte(*identity.Resources))
	if err != nil {
		return nil, err
	}

	resource := resources.GetResourceByID([]byte(resourceID))
	if resource == nil {
		return nil, errors.New("resource not found")
	}

	return resource, nil
}

// AddResource adds a new resource to the user's resources
func (r *ResourceService) AddResource(ctx context.Context, userID uid.UID, resource gwt.Resource) error {
	resources, err := r.ResourcesByIdentityID(ctx, userID)
	if err != nil {
		return err
	}

	resources.AddResource(resource)

	return r.UpdateResource(ctx, string(userID), *resources)
}

// UpdateResource adds a new resource to the user's resources
func (r *ResourceService) UpdateResource(ctx context.Context, userId string, resources gwt.Resources) error {
	identity, err := r.Repo.FindIdentityByID(ctx, userId)
	if err != nil {
		return err
	}

	res := string(resources.Serialize())
	identity.Resources = &res

	return r.Repo.Update(ctx, identity)
}

// RemoveResource removes a resource from the user's resources
func (r *ResourceService) RemoveResource(ctx context.Context, userID uid.UID, resourceID uid.UID) error {
	resources, err := r.ResourcesByIdentityID(ctx, userID)
	if err != nil {
		return err
	}

	resources.RemoveResourceByID([]byte(resourceID))

	return r.UpdateResource(ctx, string(userID), *resources)
}

// Roles ====================================================================================================

// AddRole adds a role to a specific resource
func (r *ResourceService) AddRole(ctx context.Context, userID uid.UID, resourceID uid.UID, role gwt.Role) error {
	resources, err := r.ResourcesByIdentityID(ctx, userID)
	if err != nil {
		return err
	}

	res := resources.GetResourceByID([]byte(resourceID))
	if res == nil {
		return errors.New("resource not found")
	}

	res.AssignRoles(role)
	return r.UpdateResource(ctx, string(userID), *resources)
}

// RemoveRole removes a role from a specific resource
func (r *ResourceService) RemoveRole(ctx context.Context, userID uid.UID, resourceID uid.UID, role gwt.Role) error {
	resources, err := r.ResourcesByIdentityID(ctx, userID)
	if err != nil {
		return err
	}

	res := resources.GetResourceByID([]byte(resourceID))
	if res == nil {
		return errors.New("resource not found")
	}

	res.RemoveRole(role)

	return r.UpdateResource(ctx, string(userID), *resources)
}

// UpdateRole updates a role from a specific resource
func (r *ResourceService) UpdateRole(ctx context.Context, userId, resId uid.UID, role gwt.Role) error {
	resources, err := r.ResourcesByIdentityID(ctx, userId)
	if err != nil {
		return err
	}

	res := resources.GetResourceByID([]byte(resId))
	if res == nil {
		return errors.New("resource not found")
	}

	res.RemoveRole(role)

	return r.UpdateResource(ctx, string(userId), *resources)
}

// AddClaim  adds a claim to a specific role
func (r *ResourceService) AddClaim(ctx context.Context, userID uid.UID, resourceID uid.UID, roleType gwt.RoleType, claim gwt.Claim) error {
	resources, err := r.ResourcesByIdentityID(ctx, userID)
	if err != nil {
		return err
	}

	res := resources.GetResourceByID([]byte(resourceID))
	if res == nil {
		return errors.New("resource not found")
	}

	for _, role := range res.GetRole(roleType) {
		if !role.HasClaim(claim.Key()) {
			role.AddClaim(claim.Key(), claim.Value())
			break
		}
	}

	return r.UpdateResource(ctx, string(userID), *resources)
}

// RemoveClaim removes a claim from a specific role
func (r *ResourceService) RemoveClaim(ctx context.Context, userID uid.UID, resourceID uid.UID, roleType gwt.RoleType, claim gwt.Claim) error {
	resources, err := r.ResourcesByIdentityID(ctx, userID)
	if err != nil {
		return err
	}

	res := resources.GetResourceByID([]byte(resourceID))
	if res == nil {
		return errors.New("resource not found")
	}

	for _, role := range res.GetRole(roleType) {
		if role.HasClaim(claim.Key()) {
			role.DeleteClaim(claim.Key())
			break
		}
	}

	return r.UpdateResource(ctx, string(userID), *resources)
}

// UpdateClaim updates a claim from a specific role
func (r *ResourceService) UpdateClaim(ctx context.Context, userId, resId uid.UID, roleType gwt.RoleType, claim gwt.Claim) error {
	resources, err := r.ResourcesByIdentityID(ctx, userId)
	if err != nil {
		return err
	}

	res := resources.GetResourceByID([]byte(resId))
	if res == nil {
		return errors.New("resource not found")
	}

	for _, role := range res.GetRole(roleType) {
		if role.HasClaim(claim.Key()) {
			role.ReplaceClaim(claim)
		}
	}

	return r.UpdateResource(ctx, string(userId), *resources)
}
