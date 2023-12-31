package keycloak

import (
	"context"
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

func (kc *keycloakProvider) GetUsers(ctx context.Context) ([]*gocloak.User, error) {
	return kc.SearchUsers(ctx, "")
}

func (kc *keycloakProvider) SearchUsers(ctx context.Context, filter string) ([]*gocloak.User, error) {
	users, err := kc.client.GetUsers(ctx, kc.token.AccessToken, kc.realm, gocloak.GetUsersParams{
		Search: utils.StrPtr(filter),
	})
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}

	return users, nil
}

func (kc *keycloakProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := kc.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return kc.BaseConfig.CompareUsers(sourceUsers, users, kc.BaseConfig.Mapping["id"])
}
