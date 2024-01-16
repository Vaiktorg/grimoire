package gwt

import (
	"bytes"
	"fmt"
	"github.com/vaiktorg/grimoire/uid"
	"strconv"
	"strings"
)

var (
	indexToResourceType = make(map[byte]ResourceType)
	indexToRoleType     = make(map[byte]RoleType)
	FixedIDLen          = 16
)

func init() {
	for k, v := range defaultResourceTypeIdx {
		indexToResourceType[v] = k
	}
	for k, v := range defaultRoleTypeIdx {
		indexToRoleType[v] = k
	}
}

// Resources ...
// ====================================================================================================
type Resources struct {
	UserID    []byte      // UserID: Who these resources belong to
	Resources []*Resource // Key: Resource.ResID; Value: Resource
}

func NewResources(userID uid.UID) *Resources {
	return &Resources{
		UserID: userID,
	}
}

func (res *Resources) HasAccess(resourceType ResourceType, role ...Role) bool {
	ok := 0
	for _, resource := range res.Resources {
		for _, r := range role {
			if resource.Type != resourceType {
				continue
			}

			if resource.HasRole(r) {
				ok++
			}
		}
	}

	if ok == 1 {
		return true
	}

	return false
}
func (res *Resources) String() string {
	return string(res.Serialize())
}
func (res *Resources) Serialize() []byte {
	buffer := new(bytes.Buffer)

	// Serialize UserID directly as a byte slice
	buffer.Write(res.UserID)
	fmt.Println(buffer.Bytes())

	// Serialize number of resources
	buffer.WriteByte(byte(len(res.Resources)))
	fmt.Println(buffer.Bytes())
	for _, r := range res.Resources {
		// Serialize resource ID directly as a byte slice
		buffer.Write(r.ResID)
		fmt.Println(buffer.Bytes())

		// Serialize resource type as an index
		buffer.WriteByte(resourceTypeToIndex(r.Type))
		fmt.Println(buffer.Bytes())

		// Serialize number of role assignments
		buffer.WriteByte(byte(len(r.Roles)))
		fmt.Println(buffer.Bytes())
		for _, role := range r.Roles {
			// Serialize role type
			roleTypeIndex := roleTypeToIndex(string(role.Type))
			buffer.WriteByte(roleTypeIndex)
			fmt.Println(buffer.Bytes())
			// Serialize permissions
			buffer.WriteByte(byte(role.Permissions))
			fmt.Println(buffer.Bytes())
			/// ==========
			// Len of Claims
			buffer.WriteByte(byte(len(role.Claims)))
			fmt.Println(buffer.Bytes())
			for _, claim := range role.Claims {

				// len of claim
				buffer.WriteByte(byte(len(claim)))
				fmt.Println(buffer.Bytes())
				// claim data
				buffer.Write([]byte(claim))
				fmt.Println(buffer.Bytes())
			}
		}

	}

	return buffer.Bytes()
}
func (res *Resources) Deserialize(data []byte) error {
	buffer := bytes.NewBuffer(data)
	fmt.Println(buffer.Bytes())

	// Deserialize UserID
	res.UserID = make([]byte, FixedIDLen)
	_, _ = buffer.Read(res.UserID)
	fmt.Println(buffer.Bytes())

	numResources, _ := buffer.ReadByte()
	resources := make([]Resource, numResources)

	for i := range resources {
		// Deserialize resource ID
		resources[i].ResID = make([]byte, FixedIDLen)
		_, _ = buffer.Read(resources[i].ResID)

		// Deserialize resource type
		resTypeIndex, _ := buffer.ReadByte()
		fmt.Println(buffer.Bytes())
		resources[i].Type = indexToResourceType[resTypeIndex]

		// Deserialize roles
		numRoles, _ := buffer.ReadByte()
		fmt.Println(buffer.Bytes())
		resources[i].Roles = make([]Role, numRoles)
		for j := range resources[i].Roles {
			// Deserialize role type
			roleTypeIndex, _ := buffer.ReadByte()
			fmt.Println(buffer.Bytes())
			resources[i].Roles[j].Type = indexToRoleType[roleTypeIndex]

			// Deserialize permissions
			permByte, _ := buffer.ReadByte()
			fmt.Println(buffer.Bytes())
			resources[i].Roles[j].Permissions = Permission(permByte)

			// Len Claims
			numClaims, _ := buffer.ReadByte()
			fmt.Println(buffer.Bytes())
			resources[i].Roles[j].Claims = make([]Claim, numClaims)
			for k := range resources[i].Roles[j].Claims {
				claimLen, _ := buffer.ReadByte()
				fmt.Println(buffer.Bytes())
				claim := make([]byte, claimLen)
				_, _ = buffer.Read(claim)
				fmt.Println(buffer.Bytes())
				resources[i].Roles[j].Claims[k] = Claim(claim)
			}
		}
	}

	for _, resource := range resources {
		res.Resources = append(res.Resources, &resource)
	}

	return nil
}

func (res *Resources) AddResource(nRes Resource) {
	for _, resource := range res.GetResourceByType(nRes.Type) {
		for _, role := range nRes.Roles {
			if resource.HasRole(role) {
				continue
			}

			resource.AssignRoles(role)
		}
	}
}
func (res *Resources) GetResourceByID(id []byte) *Resource {
	for _, r := range res.Resources {
		if bytes.Equal(r.ResID, id) {
			return r
		}
	}

	return nil
}
func (res *Resources) RemoveResourceByID(id []byte) {
	for idx, r := range res.Resources {
		if bytes.Equal(r.ResID, id) {
			res.Resources = append(res.Resources[:idx], res.Resources[idx+1:]...)
		}
	}
}
func (res *Resources) RemoveResourceByType(resType ResourceType) {
	for idx, r := range res.Resources {
		if r.Type == resType {
			res.Resources = append(res.Resources[:idx], res.Resources[idx+1:]...)
		}
	}
}
func (res *Resources) GetResourceByType(resType ResourceType) []*Resource {
	var byType []*Resource
	for _, r := range res.Resources {
		if r.Type == resType {
			byType = append(byType, r)
		}
	}

	return byType
}

func resourceTypeToIndex(resType ResourceType) byte {
	return defaultResourceTypeIdx[resType]
}
func roleTypeToIndex(roleType string) byte {
	return defaultRoleTypeIdx[RoleType(roleType)]
}

// Resource ...
// ====================================================================================================
type Resource struct {
	ResID []byte       // ResID: Unique ID for a resource
	Type  ResourceType // ResourceType: file,  group, api, etc.
	Roles []Role       // Roles: What can be done to this resource
}

func NewResource(name ResourceType, roles ...Role) Resource {
	id := uid.NewUID(16)
	res := Resource{
		ResID: []byte(id),
		Type:  name,
	}

	res.AssignRoles(roles...)
	return res
}
func (res *Resource) AssignRoles(roles ...Role) {
	for _, role := range roles {
		if res.HasRole(role) {
			continue
		}

		res.Roles = append(res.Roles, role)
	}
}
func (res *Resource) HasRole(role Role) bool {
	for _, r := range res.Roles {
		if role.Equals(r) {
			return true
		}
	}

	return false
}
func (res *Resource) UpdateRole(nRole Role) {
	if !res.HasRole(nRole) {
		res.Roles = append(res.Roles, nRole)
	}
}
func (res *Resource) RemoveRole(role Role) {
	for idx, r := range res.Roles {
		if role.Equals(r) {
			res.Roles = append(res.Roles[:idx], res.Roles[idx+1:]...)
			return
		}
	}
}

func (res *Resource) GetRole(nRole RoleType) []Role {
	var byType []Role
	for _, role := range res.Roles {
		if role.Type == nRole {
			byType = append(byType, role)
		}
	}

	return byType
}

// Role
// ====================================================================================================
type Role struct {
	Type        RoleType   // Type: RoleType defines the role of a user.
	Permissions Permission // Permissions: What this role is able to do as a bitwise flag. Ex: Write, Read, Edit, Delete
	Claims      []Claim    // Claims: User specific information
}

func (r *Role) HasPermission(permission Permission) bool {
	return r.Permissions&permission != 0
}
func (r *Role) ListPermissions() string {
	var permissions string
	for _, perm := range []Permission{Read, Write, Edit, Delete} {
		if r.HasPermission(perm) {
			permissions += perm.String() + ", "
		}
	}
	return permissions
}
func (r *Role) Equals(otherRole Role) bool {
	return r.Type == otherRole.Type && r.Permissions == otherRole.Permissions
}

// Claim
// --------------------------------------------------
type Claim string

func (c Claim) Key() string {
	return string(c)[strings.IndexRune(string(c), '.')-1:]
}
func (c Claim) Value() string {
	return string(c)[:strings.IndexRune(string(c), '.')+1]
}
func (c Claim) String() string {
	return strings.Replace(string(c), ".", ": ", -1)
}

// ====================================================================================================

func (r *Role) AddClaim(k string, v string) {
	if !r.HasClaim(k) {
		r.Claims = append(r.Claims, Claim(k+"."+v))
	}
}
func (r *Role) HasClaim(key string) bool {
	for _, claim := range r.Claims {
		if claim.Key() == key {
			return true
		}
	}

	return false
}
func (r *Role) GetClaim(key string) Claim {
	for _, claim := range r.Claims {
		if claim.Key() == key {
			return claim
		}
	}

	return ""
}
func (r *Role) ReplaceClaim(nClaim Claim) {
	for idx, claim := range r.Claims {
		if claim.Key() == nClaim.Key() {
			r.Claims[idx] = nClaim
			return
		}
	}
}
func (r *Role) DeleteClaim(key string) {
	for idx, claim := range r.Claims {
		if claim.Key() == key {
			r.Claims = append(r.Claims[:idx], r.Claims[idx+1:]...)
			return
		}
	}
}

func (r *Role) String() string {
	return fmt.Sprintf("Resources: %s, Permissions: %v, Claims: %v", r.Type, r.ListPermissions(), r.Claims)
}

// Permission
// ====================================================================================================
type Permission int32

const (
	Read   Permission = 1 << iota // 0001
	Write                         // 0010
	Edit                          // 0100
	Delete                        // 1000
)

func (p Permission) String() string {
	switch p {
	case Read:
		return "read"
	case Write:
		return "write"
	case Edit:
		return "edit"
	case Delete:
		return "delete"
	default:
		return strconv.Itoa(int(p))
	}
}
