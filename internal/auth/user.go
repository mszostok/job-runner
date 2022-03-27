package auth

import (
	"errors"
)

const (
	// AdminRole specified the admin role, with privileges to manage all Jobs.
	AdminRole = "admin"
	// UserRole specified the user role, with privileges to manage only owned Jobs.
	UserRole = "user"
)

// User represent Agent user entity.
type User struct {
	Name  string
	Roles map[string]struct{}
}

// NewUser returns new User instance.
func NewUser(name string, roles []string) *User {
	mapped := make(map[string]struct{}, len(roles))
	for idx := range roles {
		mapped[roles[idx]] = struct{}{}
	}

	return &User{Name: name, Roles: mapped}
}

// Validate validates if User has required properties.
func (u *User) Validate() error {
	if u == nil || u.Name == "" {
		return errors.New("user not specified")
	}

	if len(u.Roles) == 0 {
		return errors.New("user doesn't have any role assigned")
	}

	return nil
}

// CheckAuthorized checks if a given user is authorized to work with a given resource.
func (u *User) CheckAuthorized(createdBy string) error {
	if err := u.Validate(); err != nil {
		return NewGRPCInvalidCertError(err)
	}

	// only admin may work with not owned resources
	if u.Name != createdBy && !u.hasRole(AdminRole) {
		return NewGRPCPermissionDeniedError()
	}

	if !u.hasAnyRole(UserRole, AdminRole) {
		return NewGRPCPermissionDeniedError()
	}

	return nil
}

func (u *User) hasRole(exp string) bool {
	_, found := u.Roles[exp]
	return found
}

func (u *User) hasAnyRole(exp ...string) bool {
	for _, role := range exp {
		if u.hasRole(role) {
			return true
		}
	}
	return false
}
