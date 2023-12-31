package provider

import "context"

// type IdentityProvider interface {
// 	GetUsers(context.Context) ([]map[string]any, error)
// 	SearchUsers(context.Context, string) ([]map[string]any, error)
// 	GetUser(context.Context, string) (map[string]any, error)
// 	// GetGroups(context.Context)
// 	Sync(context.Context)
// 	// GetMapping() map[string]string
// 	ConvertUser(user any) (map[string]any, error)
// 	CompareUsers(source []map[string]any, target []map[string]any, field string) (toAdd, toRemove, toUpdate []map[string]any, retErr error)
// }

// type SyncIdentityProvider interface {
// 	GetUsers(context.Context) ([]map[string]any, error)
// 	SearchUsers(context.Context, string) ([]map[string]any, error)
// 	GetUser(context.Context, string) (map[string]any, error)

// }

type IdentityProvider interface {
	Sync(ctx context.Context, target []map[string]any) (toAdd, toRemove, toUpdate []map[string]any, retErr error)
}

type ProviderOption func()

// type UserMapper struct{}
// func WithUserMapping(map[string]string) func(IdentityProvider)
