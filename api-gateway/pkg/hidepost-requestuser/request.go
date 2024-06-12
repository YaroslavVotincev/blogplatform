package requestuser

const (
	UserRoleUnknown   = "unknown"
	UserRoleUser      = "user"
	UserRoleModerator = "moderator"
	UserRoleAdmin     = "admin"
	UserRoleService   = "service"

	UserRoleHeaderKey     = "USER-ROLE"
	UserIdHeaderKey       = "USER-ID"
	UserIsBannedHeaderKey = "USER-BANNED"

	AuthorizationHeaderKey = "Authorization"
)
