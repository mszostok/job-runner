package auth

import (
	"context"
)

// key is an unexported type for keys defined in this package.
// This prevents collisions with keys defined in other packages.
type key int

// userKey is the key for user.User values in Contexts. It is unexported.
// Clients use user.NewContext and user.FromContext instead of using this key directly.
var userKey key

// NewContext returns a new Context with a given user.
func NewContext(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// FromContext returns the User stored in ctx, or error if not available.
func FromContext(ctx context.Context) (*User, error) {
	u, ok := ctx.Value(userKey).(*User)
	if !ok || u == nil {
		return nil, NewGRPCMissingCertError()
	}
	return u, nil
}
