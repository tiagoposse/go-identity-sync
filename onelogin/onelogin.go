package onelogin

import (
	"context"
	"fmt"
	"strconv"

	"github.com/onelogin/onelogin-go-sdk/v4/pkg/onelogin"
	"github.com/onelogin/onelogin-go-sdk/v4/pkg/onelogin/models"
	"github.com/tiagoposse/go-identity-sync/config"
)

type oneloginProvider struct {
	config.BaseConfig

	client *onelogin.OneloginSDK
}

func NewOneloginProvider(ctx context.Context, cfg config.OneLoginConfig) (*oneloginProvider, error) {
	ol, err := onelogin.NewOneloginSDK()
	if err != nil {
		return nil, fmt.Errorf("initialize client: %w", err)
	}

	return &oneloginProvider{
		BaseConfig: cfg.BaseConfig,
		client:     ol,
	}, nil
}

func (ol *oneloginProvider) GetUser(ctx context.Context, id string) (*models.User, error) {
	numID, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := ol.client.GetUserByID(numID, nil)
	if err != nil {
		fmt.Printf("Error Getting User: %v\n", err)
	}

	return user.(*models.User), nil
}

func (ol *oneloginProvider) GetUsers(ctx context.Context) ([]*models.User, error) {
	return ol.SearchUsers(ctx, "")
}

func (ol *oneloginProvider) SearchUsers(ctx context.Context, filter string) ([]*models.User, error) {
	userQuery := models.UserQuery{}

	userList, err := ol.client.GetUsers(&userQuery)
	if err != nil {
		return nil, fmt.Errorf("searching users: %w", err)
	}

	return userList.([]*models.User), nil
}

func (ol *oneloginProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := ol.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return ol.BaseConfig.CompareUsers(sourceUsers, users, ol.BaseConfig.Mapping["id"])
}
