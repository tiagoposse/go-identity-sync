package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v57/github"
	"github.com/tiagoposse/go-identity-sync/config"
)

type githubProvider struct {
	config.BaseConfig

	client *github.Client
	org    string
}

func NewgithubProvider(ctx context.Context, cfg *config.GithubConfig) (*githubProvider, error) {
	client := github.NewClient(nil).WithAuthToken(*cfg.Token.Value)

	return &githubProvider{
		client:     client,
		org:        cfg.Organisation,
		BaseConfig: cfg.BaseConfig,
	}, nil
}

func (gh *githubProvider) GetUser(ctx context.Context, id string) (*github.User, error) {
	user, _, err := gh.client.Users.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("fetching user details: %w", err)
	}

	return user, nil
}

func (gh *githubProvider) GetUsers(ctx context.Context) ([]*github.User, error) {
	return gh.SearchUsers(ctx, "all")
}

func (gh *githubProvider) SearchUsers(ctx context.Context, filter string) ([]*github.User, error) {
	users, _, err := gh.client.Organizations.ListMembers(ctx, gh.org, &github.ListMembersOptions{
		Filter: filter,
	})

	if err != nil {
		return nil, fmt.Errorf("fetching members: %w", err)
	}

	return users, nil
}

func (gh *githubProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := gh.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return gh.BaseConfig.CompareUsers(sourceUsers, users, gh.BaseConfig.Mapping["id"])
}
