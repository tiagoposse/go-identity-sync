package google

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tiagoposse/go-identity-sync/config"
	"github.com/tiagoposse/go-identity-sync/utils"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

type googleProvider struct {
	config.BaseConfig

	client *admin.Service
	domain string
}

func NewGoogleProvider(ctx context.Context, cfg *config.GoogleConfig) (*googleProvider, error) {
	// Configure the JWT config
	gcfg, err := google.JWTConfigFromJSON(
		[]byte(*cfg.ServiceAccountKey.Value),
		admin.AdminDirectoryUserReadonlyScope,
		admin.AdminDirectoryGroupMemberReadonlyScope,
		admin.AdminDirectoryGroupReadonlyScope,
	)

	gcfg.Subject = cfg.UserToImpersonate
	if err != nil {
		return nil, fmt.Errorf("creating google config: %w", err)
	}

	adminService, err := admin.NewService(ctx, option.WithHTTPClient(gcfg.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("creating google admin client: %w", err)
	}

	return &googleProvider{
		BaseConfig: cfg.BaseConfig,
		client:     adminService,
		domain:     cfg.Domain,
	}, nil
}

func (gac *googleProvider) GetUser(ctx context.Context, id string) (*admin.User, error) {
	user, err := gac.client.Users.Get(id).Do()
	if err != nil {
		return nil, fmt.Errorf("fetching user details: %w", err)
	}

	return user, nil
}

func (gac *googleProvider) GetUsers(ctx context.Context, lo utils.ListOptions) ([]*admin.User, error) {
	query := gac.client.Users.List().Domain(gac.domain)
	if lo.Filter != nil {
		query.Query(*lo.Filter)
	}

	res, err := query.Do()
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}
	return res.Users, nil
}

func (gac *googleProvider) GetUsersAndMemberships(ctx context.Context, lo utils.ListOptions) ([]*admin.User, map[string][]string, error) {
	users, err := gac.GetUsers(ctx, lo)
	if err != nil {
		return nil, nil, err
	}
	groupMemberships := make(map[string][]string)

	for _, user := range users {
		res, err := gac.client.Groups.List().Domain(gac.domain).UserKey(user.Id).Do()
		if err != nil {
			return nil, nil, fmt.Errorf("getting groups for user %s: %w", user.PrimaryEmail, err)
		}

		groupMemberships[user.Id] = make([]string, 0)
		for _, g := range res.Groups {
			groupMemberships[user.Id] = append(groupMemberships[user.Id], g.Email)
		}
	}

	return users, groupMemberships, nil
}

func (gac *googleProvider) GetUsersConverted(ctx context.Context, lo utils.ListOptions) ([]map[string]any, error) {
	users, memberships, err := gac.GetUsersAndMemberships(ctx, lo)
	if err != nil {
		return nil, err
	}

	convertedUsers := make([]map[string]any, 0)
	for _, item := range users {
		if user, err := gac.BaseConfig.ConvertUser(item); err != nil {
			return nil, err
		} else {
			user[gac.BaseConfig.GroupField] = memberships[item.Id]
			convertedUsers = append(convertedUsers, user)
		}

	}
	return convertedUsers, nil
}

func (gac *googleProvider) SyncProvider(ctx context.Context, source []map[string]any) error {
	targetUsers, err := gac.GetUsers(ctx, utils.ListOptions{})
	if err != nil {
		return err
	}

	toAdd, toRemove, toUpdate, err := gac.BaseConfig.CompareUsers(source, targetUsers, gac.BaseConfig.Mapping["id"])
	if err != nil {
		return err
	}

	for _, u := range toAdd {
		mapped, err := gac.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		conv, err := MapToUser(mapped)
		if err != nil {
			return err
		}

		if _, err := gac.client.Users.Insert(conv).Do(); err != nil {
			return err
		}
	}

	for _, u := range toRemove {
		if err := gac.client.Users.Delete(u[gac.BaseConfig.Mapping["id"]].(string)).Do(); err != nil {
			return err
		}
	}

	for _, u := range toUpdate {
		mapped, err := gac.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		conv, err := MapToUser(mapped)
		if err != nil {
			return err
		}

		if _, err := gac.client.Users.Update(
			conv.Id, conv,
		).Do(); err != nil {
			return err
		}
	}

	return nil
}

// func (gac *googleProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
// 	sourceUsers, err := gac.GetUsersRaw(ctx)
// 	if err != nil {
// 		retErr = err
// 		return
// 	}

// 	return gac.BaseConfig.CompareUsers(sourceUsers, users, gac.BaseConfig.Mapping["id"])
// }

func MapToUser(user map[string]any) (*admin.User, error) {
	bs, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	var original *admin.User
	if err := json.Unmarshal(bs, &original); err != nil {
		return nil, err
	}

	return original, nil
}
