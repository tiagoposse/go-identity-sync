package okta

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/tiagoposse/go-identity-sync/config"
	"github.com/tiagoposse/go-identity-sync/utils"
)

type oktaProvider struct {
	config.BaseConfig

	client *okta.Client
}

func NewOktaProvider(ctx context.Context, cfg *config.OktaConfig) (*oktaProvider, error) {
	_, cli, err := okta.NewClient(
		ctx,
		okta.WithOrgUrl(cfg.Domain),
		okta.WithToken(*cfg.Token.Value),
	)

	return &oktaProvider{
		client:     cli,
		BaseConfig: cfg.BaseConfig,
	}, err
}

func (ok *oktaProvider) GetUser(ctx context.Context, id string) {
	user, resp, err := ok.client.User.GetUser(ctx, id)
	if err != nil {
		fmt.Printf("Error Getting User: %v\n", err)
	}
	fmt.Printf("User: %+v\n Response: %+v\n\n", user, resp)
}

func (ok *oktaProvider) GetUsers(ctx context.Context, lo utils.ListOptions) ([]*okta.User, error) {
	fquery := query.NewQueryParams(query.WithFilter(*lo.Filter))

	users, _, err := ok.client.User.ListUsers(ctx, fquery)
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}

	return users, nil
}

func (ok *oktaProvider) GetUsersAndMemberships(ctx context.Context, lo utils.ListOptions) ([]*okta.User, map[string][]string, error) {
	users, err := ok.GetUsers(ctx, lo)
	if err != nil {
		return nil, nil, fmt.Errorf("getting users and memberships: %w", err)
	}

	memberships := make(map[string][]string)
	for _, user := range users {
		groups, _, err := ok.client.User.ListUserGroups(ctx, user.Id)
		if err != nil {
			return nil, nil, err
		}
		memberships[user.Id] = make([]string, 0)
		for _, g := range groups {
			memberships[user.Id] = append(memberships[user.Id], g.Id)
		}
	}

	return users, memberships, nil
}

func (ok *oktaProvider) GetUsersConverted(ctx context.Context, lo utils.ListOptions) ([]map[string]any, error) {
	users, memberships, err := ok.GetUsersAndMemberships(ctx, lo)
	if err != nil {
		return nil, err
	}

	convertedUsers := make([]map[string]any, 0)
	for _, item := range users {
		if user, err := ok.BaseConfig.ConvertUser(item); err != nil {
			return nil, err
		} else {
			user[ok.BaseConfig.GroupField] = memberships[item.Id]
			convertedUsers = append(convertedUsers, user)
		}

	}
	return convertedUsers, nil
}

func (ok *oktaProvider) SyncProvider(ctx context.Context, source []map[string]any) error {
	targetUsers, err := ok.GetUsers(ctx, utils.ListOptions{})
	if err != nil {
		return err
	}

	toAdd, toRemove, toUpdate, err := ok.BaseConfig.CompareUsers(source, targetUsers, ok.BaseConfig.Mapping["id"])
	if err != nil {
		return err
	}

	for _, u := range toAdd {
		mapped, err := ok.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		cur := okta.CreateUserRequest{
			Type: &okta.UserType{
				DisplayName: mapped[ok.BaseConfig.Mapping["type.displayName"]].(string),
				Name:        mapped[ok.BaseConfig.Mapping["type.name"]].(string),
				Description: mapped[ok.BaseConfig.Mapping["type.description"]].(string),
			},
			GroupIds: u[ok.BaseConfig.GroupField].([]string),
		}

		if _, _, err := ok.client.User.CreateUser(ctx, cur, nil); err != nil {
			return err
		}
	}

	for _, u := range toRemove {
		if _, err := ok.client.User.DeactivateOrDeleteUser(ctx, u[ok.BaseConfig.Mapping["id"]].(string), nil); err != nil {
			return err
		}
	}

	for _, u := range toUpdate {
		mapped, err := ok.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		conv, err := MapToUser(mapped)
		if err != nil {
			return err
		}

		if _, _, err := ok.client.User.UpdateUser(
			ctx,
			u[ok.BaseConfig.Mapping["id"]].(string),
			*conv, nil,
		); err != nil {
			return err
		}
	}

	return nil
}

func (ok *oktaProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := ok.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return ok.BaseConfig.CompareUsers(sourceUsers, users, ok.BaseConfig.Mapping["id"])
}

func MapToUser(user map[string]any) (*okta.User, error) {
	bs, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	var original *okta.User
	if err := json.Unmarshal(bs, &original); err != nil {
		return nil, err
	}

	return original, nil
}
