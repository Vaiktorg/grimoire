package gwt

// ResourceType ====================================================================================================
type ResourceType string

const (
	Network         ResourceType = "network"
	DataManagement               = "data_management"
	UserInterface                = "user_interface"
	SecurityMonitor              = "security_monitor"
	SystemAdmin                  = "system_admin"
	DevTools                     = "dev_tools"
	ThirdParty                   = "third_party"
)

func (r ResourceType) String() string {
	return string(r)
}

var defaultResourceTypeIdx = map[ResourceType]int{}

// Roles ====================================================================================================

type RoleType string

func (r RoleType) String() string {
	return string(r)
}

// RoleType default constants.
const (
	Owner RoleType = "own" // Owner (Company)
	Admin          = "adm" // Administrator
	Dev            = "dev" // Developer
	Mod            = "mod" // Moderator
	Guest          = "gst" // Guest
	User           = "usr" // User
)

func (r RoleType) Hierarchy(compare RoleType) bool {
	t1, ok1 := defaultRoleTypeIdx[r]
	t2, ok2 := defaultRoleTypeIdx[compare]

	if !ok1 || !ok2 {
		return false
	}

	return t1 <= t2
}

// DefaultRoles ...
var DefaultRoles = map[RoleType]Role{
	Owner: {Type: Owner, Permissions: Read | Write | Edit | Delete},
	Admin: {Type: Admin, Permissions: Read | Write | Edit | Delete},
	Dev:   {Type: Dev, Permissions: Read | Write | Edit | Delete},
	Mod:   {Type: Mod, Permissions: Read | Edit},
	Guest: {Type: Guest, Permissions: Read},
	User:  {Type: User, Permissions: Read},
}

var defaultRoleTypeIdx = map[RoleType]int{
	Owner: 0,
	Admin: 1,
	Dev:   2,
	Mod:   3,
	Guest: 4,
	User:  5,
}
