package provider

import (
	"context"
)

type SyncProvider struct{}

func Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	currentUsers, err := p.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	sourceUsers := make([]map[string]any, 0)
	for _, item := range currentUsers {
		if user, err := p.ConvertUser(item); err != nil {
			retErr = err
			return
		} else {
			sourceUsers = append(sourceUsers, user)
		}
	}

	return p.CompareUsers(sourceUsers, users, p.Mapping["UserId"])
}
