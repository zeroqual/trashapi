package helpers

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type Permission string

const (
	UserUpdate Permission = "user.update"
)

var rolePermissions = map[Role]map[Permission]bool{
	RoleAdmin: {
		UserUpdate: true,
	},

	RoleUser: {
		UserUpdate: true,
	},
}

func HasPremission(role Role, permission Permission) bool {
	permissions, ok := rolePermissions[role]
	if !ok {
		return false
	}

	return permissions[permission]
}
