package gwt

// ResourceType ====================================================================================================
type ResourceType string

const (
	NetworkPublicEndpoints   ResourceType = "networkPublicEndpoints"
	NetworkAPIGateway        ResourceType = "networkAPIGateway"
	NetworkWebhooks          ResourceType = "networkWebhooks"
	NetworkInternalEndpoints ResourceType = "networkInternalEndpoints"
	NetworkDatabaseAPI       ResourceType = "networkDatabaseAPI"
	NetworkInternalServices  ResourceType = "networkInternalServices"

	DataManagementDatabases         ResourceType = "dataManagementDatabases"
	DataManagementUserData          ResourceType = "dataManagementUserData"
	DataManagementAnalyticsData     ResourceType = "dataManagementAnalyticsData"
	DataManagementFileStorage       ResourceType = "dataManagementFileStorage"
	DataManagementMedia             ResourceType = "dataManagementMedia"
	DataManagementDocumentsArchives ResourceType = "dataManagementDocumentsArchives"

	UIAdminDashboard        ResourceType = "uiAdminDashboard"
	UIConfigurationSettings ResourceType = "uiConfigurationSettings"
	UIUserManagement        ResourceType = "uiUserManagement"
	UIPublicWebsite         ResourceType = "uiPublicWebsite"
	UIHomePage              ResourceType = "uiHomePage"
	UIContactForm           ResourceType = "uiContactForm"
	UIBlogArticles          ResourceType = "uiBlogArticles"

	SecMonitorFirewallSettings ResourceType = "secMonitorFirewallSettings"
	SecMonitorAccessLogs       ResourceType = "secMonitorAccessLogs"
	SecMonitorSecurityAlerts   ResourceType = "secMonitorSecurityAlerts"

	SysAdminServerManagement     ResourceType = "sysAdminServerManagement"
	SysAdminVirtualMachines      ResourceType = "sysAdminVirtualMachines"
	SysAdminContainerInstances   ResourceType = "sysAdminContainerInstances"
	SysAdminNetworkConfiguration ResourceType = "sysAdminNetworkConfiguration"
	SysAdminDNSSettings          ResourceType = "sysAdminDNSSettings"
	SysAdminSSLCertificates      ResourceType = "sysAdminSSLCertificates"

	DevToolsSourceCodeRepositories     ResourceType = "devToolsSourceCodeRepositories"
	DevToolsCICD                       ResourceType = "devToolsCICD"
	DevToolsTestingStagingEnvironments ResourceType = "devToolsTestingStagingEnvironments"

	ThirdPartyCloudServiceProviders ResourceType = "thirdPartyCloudServiceProviders"
	ThirdPartyExternalAPIs          ResourceType = "thirdPartyExternalAPIs"
	ThirdPartyIntegrationPlatforms  ResourceType = "thirdPartyIntegrationPlatforms"
)

var defaultResourceTypeIdx = map[ResourceType]byte{
	NetworkPublicEndpoints:   0x01,
	NetworkAPIGateway:        0x02,
	NetworkWebhooks:          0x03,
	NetworkInternalEndpoints: 0x04,
	NetworkDatabaseAPI:       0x05,
	NetworkInternalServices:  0x06,

	DataManagementDatabases:         0x07,
	DataManagementUserData:          0x08,
	DataManagementAnalyticsData:     0x09,
	DataManagementFileStorage:       0x10,
	DataManagementMedia:             0x11,
	DataManagementDocumentsArchives: 0x12,

	UIAdminDashboard:        0x13,
	UIConfigurationSettings: 0x14,
	UIUserManagement:        0x15,
	UIPublicWebsite:         0x16,
	UIHomePage:              0x17,
	UIContactForm:           0x18,
	UIBlogArticles:          0x19,

	SecMonitorFirewallSettings: 0x20,
	SecMonitorAccessLogs:       0x21,
	SecMonitorSecurityAlerts:   0x22,

	SysAdminServerManagement:     0x23,
	SysAdminVirtualMachines:      0x24,
	SysAdminContainerInstances:   0x25,
	SysAdminNetworkConfiguration: 0x26,
	SysAdminDNSSettings:          0x27,
	SysAdminSSLCertificates:      0x28,

	DevToolsSourceCodeRepositories:     0x29,
	DevToolsCICD:                       0x30,
	DevToolsTestingStagingEnvironments: 0x31,

	ThirdPartyCloudServiceProviders: 0x32,
	ThirdPartyExternalAPIs:          0x33,
	ThirdPartyIntegrationPlatforms:  0x34,
}

// Roles ====================================================================================================

type RoleType string

// RoleType default constants.
const (
	Owner RoleType = "owner" // Owner (Company)
	Admin          = "admin" // Administrator
	Dev            = "dev"   // Developer
	Mod            = "mod"   // Moderator
	Guest          = "guest" // Guest
	User           = "user"  // User
)

func (rt RoleType) Hierarchy(compare RoleType) bool {
	t1, ok1 := defaultRoleTypeIdx[rt]
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

var defaultRoleTypeIdx = map[RoleType]byte{
	Owner: 0x01,
	Admin: 0x02,
	Dev:   0x03,
	Mod:   0x04,
	Guest: 0x05,
	User:  0x06,

	// Add more role types as needed
}
