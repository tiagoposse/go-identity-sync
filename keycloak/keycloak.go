package keycloak

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Nerzal/gocloak/v13"
	"github.com/tiagoposse/go-identity-sync/config"
	"github.com/tiagoposse/go-identity-sync/utils"
)

type keycloakProvider struct {
	config.BaseConfig
	client *gocloak.GoCloak
	realm  string

	token *gocloak.JWT
}

func NewKeycloakProvider(ctx context.Context, cfg *config.KeycloakConfig) (*keycloakProvider, error) {
	client := gocloak.NewClient(cfg.Url)
	token, err := client.LoginAdmin(ctx, *cfg.Username.Value, *cfg.Password.Value, cfg.Realm)
	if err != nil {
		return nil, errors.New("something wrong with the credentials or url")
	}

	return &keycloakProvider{
		BaseConfig: cfg.BaseConfig,
		client:     client,
		realm:      cfg.Realm,
		token:      token,
	}, nil
}

func (kc *keycloakProvider) GetUser(ctx context.Context, id string) (*gocloak.User, error) {
	user, err := kc.client.GetUserByID(ctx, kc.token.AccessToken, kc.realm, id)
	if err != nil {
		return nil, fmt.Errorf("fetching user details: %w", err)
	}

	return user, nil
}

func (kc *keycloakProvider) GetUsersConverted(ctx context.Context) ([]map[string]any, error) {
	users, memberships, err := kc.GetUsersAndMemberships(ctx, "")
	if err != nil {
		return nil, err
	}

	convertedUsers := make([]map[string]any, 0)
	for _, item := range users {
		if user, err := kc.BaseConfig.ConvertUser(item); err != nil {
			return nil, err
		} else {
			user[kc.BaseConfig.GroupField] = memberships[string(*item.ID)]
			convertedUsers = append(convertedUsers, user)
		}
	}

	return convertedUsers, nil
}

func (kc *keycloakProvider) GetUsers(ctx context.Context, lo utils.ListOptions) ([]*gocloak.User, error) {
	return kc.client.GetUsers(ctx, kc.token.AccessToken, kc.realm, gocloak.GetUsersParams{
		Search: lo.Filter,
	})
}

func (kc *keycloakProvider) SyncProvider(ctx context.Context, source []map[string]any) error {
	targetUsers, err := kc.GetUsers(ctx, utils.ListOptions{})
	if err != nil {
		return err
	}

	toAdd, toRemove, toUpdate, err := kc.BaseConfig.CompareUsers(source, targetUsers, kc.BaseConfig.Mapping["id"])
	if err != nil {
		return err
	}

	for _, u := range toAdd {
		mapped, err := kc.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		conv, err := MapToUser(mapped)
		if err != nil {
			return err
		}

		if _, err := kc.client.CreateUser(ctx, kc.token.AccessToken, kc.realm, *conv); err != nil {
			return err
		}
	}

	for _, u := range toRemove {
		if err := kc.client.DeleteUser(ctx, kc.token.AccessToken, kc.realm, u[kc.BaseConfig.Mapping["id"]].(string), nil); err != nil {
			return err
		}
	}

	for _, u := range toUpdate {
		mapped, err := kc.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		conv, err := MapToUser(mapped)
		if err != nil {
			return err
		}

		if err := kc.client.UpdateUser(
			ctx,
			kc.token.AccessToken, kc.realm,
			*conv,
		); err != nil {
			return err
		}
	}

	return nil
}

func (kc *keycloakProvider) GetUsersAndMemberships(ctx context.Context, filter string) ([]*gocloak.User, map[string][]string, error) {
	users, err := kc.client.GetUsers(ctx, kc.token.AccessToken, kc.realm, gocloak.GetUsersParams{
		Search: utils.StrPtr(filter),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("getting users: %w", err)
	}

	memberships := make(map[string][]string)
	for _, u := range users {
		memberships[*u.ID] = *u.Groups
	}

	return users, memberships, nil
}

func (kc *keycloakProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := kc.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return kc.BaseConfig.CompareUsers(sourceUsers, users, kc.BaseConfig.Mapping["id"])
}

func MapToUser(user map[string]any) (*gocloak.User, error) {
	bs, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	var original *gocloak.User
	if err := json.Unmarshal(bs, &original); err != nil {
		return nil, err
	}

	return original, nil
}
