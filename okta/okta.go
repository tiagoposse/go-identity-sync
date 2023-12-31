package okta

import (
	"context"
	"fmt"

	"github.com/okta/okta-sdk-golang/v2/okta"
	"github.com/okta/okta-sdk-golang/v2/okta/query"
	"github.com/tiagoposse/go-identity-sync/config"
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

func (ok *oktaProvider) GetUsers(ctx context.Context) ([]*okta.User, error) {
	return ok.SearchUsers(ctx, "")
}

func (ok *oktaProvider) SearchUsers(ctx context.Context, filter string) ([]*okta.User, error) {
	fquery := query.NewQueryParams(query.WithFilter(filter))

	filteredUsers, resp, err := ok.client.User.ListUsers(ctx, fquery)
	if err != nil {
		fmt.Printf("Error Getting Users: %v\n", err)
	}

	fmt.Printf("Filtered Users: %+v\n Response: %+v\n\n", filteredUsers, resp)

	for index, user := range filteredUsers {
		fmt.Printf("User %d: %+v\n", index, user)
	}

	return filteredUsers, nil
}

func (ok *oktaProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := ok.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return ok.BaseConfig.CompareUsers(sourceUsers, users, ok.BaseConfig.Mapping["id"])
}
