package google

import (
	"context"
	"fmt"

	"github.com/tiagoposse/go-identity-sync/config"
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

func (gac *googleProvider) GetUsers(ctx context.Context) ([]*admin.User, error) {
	return gac.SearchUsers(ctx, "")
}

func (gac *googleProvider) SearchUsers(ctx context.Context, filter string) ([]*admin.User, error) {
	res, err := gac.client.Users.List().Domain(gac.domain).Query(filter).Do()
	if err != nil {
		return nil, fmt.Errorf("getting users: %w", err)
	}

	return res.Users, nil
}

func (gac *googleProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := gac.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return gac.BaseConfig.CompareUsers(sourceUsers, users, gac.BaseConfig.Mapping["id"])
}
