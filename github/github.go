package github

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/go-github/v57/github"
	"github.com/tiagoposse/go-identity-sync/config"
	"github.com/tiagoposse/go-identity-sync/utils"
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

func (gh *githubProvider) GetUsersAndMemberships(ctx context.Context, filter string) ([]*github.User, map[string][]string, error) {
	users, _, err := gh.client.Organizations.ListMembers(ctx, gh.org, &github.ListMembersOptions{
		Filter: filter,
	})

	if err != nil {
		return nil, nil, fmt.Errorf("fetching members: %w", err)
	}

	teams, _, err := gh.client.Teams.ListTeams(ctx, gh.org, &github.ListOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("fetching teams: %w", err)
	}

	groupMemberships := make(map[string][]string)
	for _, team := range teams {
		members, _, err := gh.client.Teams.ListTeamMembersBySlug(ctx, gh.org, *team.Slug, &github.TeamListTeamMembersOptions{})
		if err != nil {
			return nil, nil, fmt.Errorf("fetching team members for %d: %w", *team.ID, err)
		}

		for _, member := range members {
			for _, user := range users {
				uID := string(*user.ID)
				if *user.ID == *member.ID {
					if _, ok := groupMemberships[uID]; !ok {
						groupMemberships[uID] = []string{string(*team.ID)}
					} else {
						groupMemberships[uID] = append(groupMemberships[uID], string(*team.ID))
					}
				}
			}
		}
	}

	return users, groupMemberships, nil
}

func (gh *githubProvider) GetUsersConverted(ctx context.Context) ([]map[string]any, error) {
	users, memberships, err := gh.GetUsersAndMemberships(ctx, "")
	if err != nil {
		return nil, err
	}

	convertedUsers := make([]map[string]any, 0)
	for _, item := range users {
		if user, err := gh.BaseConfig.ConvertUser(item); err != nil {
			return nil, err
		} else {
			user[gh.BaseConfig.GroupField] = memberships[string(*item.ID)]
			convertedUsers = append(convertedUsers, user)
		}
	}

	return convertedUsers, nil
}

func (gh *githubProvider) GetUsers(ctx context.Context, lo utils.ListOptions) ([]*github.User, error) {
	opts := &github.ListMembersOptions{}
	if lo.Filter != nil {
		opts.Filter = *lo.Filter
	}
	users, _, err := gh.client.Organizations.ListMembers(ctx, gh.org, opts)

	if err != nil {
		return nil, fmt.Errorf("fetching members: %w", err)
	}

	return users, nil
}

func (gh *githubProvider) SyncProvider(ctx context.Context, source []map[string]any) error {
	targetUsers, err := gh.GetUsers(ctx, utils.ListOptions{})
	if err != nil {
		return err
	}

	toAdd, toRemove, toUpdate, err := gh.BaseConfig.CompareUsers(source, targetUsers, gh.BaseConfig.Mapping["id"])
	if err != nil {
		return err
	}

	for _, u := range toAdd {
		mapped, err := gh.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		iID, err := strconv.Atoi(mapped[gh.BaseConfig.Mapping["id"]].(string))
		if err != nil {
			return err
		}
		if _, _, err := gh.client.Organizations.CreateOrgInvitation(ctx, gh.org, &github.CreateOrgInvitationOptions{
			InviteeID: utils.Int64Ptr(int64(iID)),
		}); err != nil {
			return err
		}
	}

	for _, u := range toRemove {
		if _, err := gh.client.Organizations.RemoveMember(ctx, gh.org, u[gh.BaseConfig.Mapping["id"]].(string)); err != nil {
			return err
		}
	}

	for _, u := range toUpdate {
		mapped, err := gh.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		conv, err := MapToUser(mapped)
		if err != nil {
			return err
		}

		if _, _, err := gh.client.User.UpdateUser(
			ctx,
			u[gh.BaseConfig.Mapping["id"]].(string),
			*conv, nil,
		); err != nil {
			return err
		}
	}

	return nil
}

func (gh *githubProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := gh.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return gh.BaseConfig.CompareUsers(sourceUsers, users, gh.BaseConfig.Mapping["id"])
}

func MapToUser(user map[string]any) (*github.User, error) {
	bs, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	var original *github.User
	if err := json.Unmarshal(bs, &original); err != nil {
		return nil, err
	}

	return original, nil
}
