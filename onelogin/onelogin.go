package onelogin

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/onelogin/onelogin-go-sdk/v4/pkg/onelogin"
	"github.com/onelogin/onelogin-go-sdk/v4/pkg/onelogin/models"
	utl "github.com/onelogin/onelogin-go-sdk/v4/pkg/onelogin/utilities"
	"github.com/tiagoposse/go-identity-sync/config"
	"github.com/tiagoposse/go-identity-sync/utils"
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

func (ol *oneloginProvider) GetUsers(ctx context.Context, lo utils.ListOptions) ([]*models.User, error) {
	userQuery := models.UserQuery{}

	resp, err := ol.client.GetUsers(&userQuery)
	if err != nil {
		return nil, fmt.Errorf("searching users: %w", err)
	}

	var users []*models.User
	if lo.Filter != nil {
		users = make([]*models.User, 0)
		for _, u := range resp.([]*models.User) {
			if strings.Contains(u.Email, *lo.Filter) ||
				strings.Contains(u.Firstname, *lo.Filter) ||
				strings.Contains(u.Lastname, *lo.Filter) ||
				strings.Contains(u.Username, *lo.Filter) ||
				strings.Contains(u.UserPrincipalName, *lo.Filter) ||
				strings.Contains(u.Phone, *lo.Filter) ||
				strings.Contains(u.Title, *lo.Filter) ||
				strings.Contains(string(u.GroupID), *lo.Filter) ||
				strings.Contains(string(u.ID), *lo.Filter) ||
				strings.Contains(string(u.ExternalID), *lo.Filter) {

				users = append(users, u)
			}
		}
	} else {
		users = resp.([]*models.User)
	}
	return users, nil
}

func (ol *oneloginProvider) GetGroups(ctx context.Context, lo utils.ListOptions) ([]*models.Group, error) {
	resp, err := ol.client.GetGroups()
	if err != nil {
		return nil, fmt.Errorf("searching groups: %w", err)
	}

	var groups []*models.Group
	if lo.Filter != nil {
		groups = make([]*models.Group, 0)
		for _, g := range resp.([]*models.Group) {
			if strings.Contains(g.Name, *lo.Filter) ||
				strings.Contains(string(g.ID), *lo.Filter) ||
				strings.Contains(string(*g.Reference), *lo.Filter) {

				groups = append(groups, g)
			}
		}
	} else {
		groups = resp.([]*models.Group)
	}
	return groups, nil
}
func (ol *oneloginProvider) GetUsersAndMemberships(ctx context.Context, lo utils.ListOptions) ([]*models.User, map[string][]string, error) {
	users, err := ol.GetUsers(ctx, lo)
	if err != nil {
		return nil, nil, err
	}

	memberships := make(map[string][]string)
	for _, user := range users {
		uID := string(user.ID)
		resp, err := ol.client.Client.Get(utils.StrPtr(fmt.Sprintf("/api/2/users/%s", uID)), nil)
		if err != nil {
			return nil, nil, err
		}
		groups, err := utl.CheckHTTPResponse(resp)
		if err != nil {
			return nil, nil, err
		}
		memberships[uID] = make([]string, 0)
		for _, g := range groups.([]models.Group) {
			memberships[uID] = append(memberships[uID], string(g.ID))
		}
	}

	return users, memberships, nil
}

func (ol *oneloginProvider) SearchUsers(ctx context.Context, filter string) ([]*models.User, error) {
	userQuery := models.UserQuery{}

	userList, err := ol.client.GetUsers(&userQuery)
	if err != nil {
		return nil, fmt.Errorf("searching users: %w", err)
	}

	return userList.([]*models.User), nil
}

// func (ol *oneloginProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
// 	sourceUsers, err := ol.GetUsers(ctx, )
// 	if err != nil {
// 		retErr = err
// 		return
// 	}

// 	return ol.BaseConfig.CompareUsers(sourceUsers, users, ol.BaseConfig.Mapping["id"])
// }

func (ol *oneloginProvider) GetUsersConverted(ctx context.Context, lo utils.ListOptions) ([]map[string]any, error) {
	users, memberships, err := ol.GetUsersAndMemberships(ctx, lo)
	if err != nil {
		return nil, err
	}

	convertedUsers := make([]map[string]any, 0)
	for _, item := range users {
		if user, err := ol.BaseConfig.ConvertUser(item); err != nil {
			return nil, err
		} else {
			user[ol.BaseConfig.GroupField] = memberships[string(item.ID)]
			convertedUsers = append(convertedUsers, user)
		}

	}
	return convertedUsers, nil
}

func (ol *oneloginProvider) SyncProvider(ctx context.Context, source []map[string]any) error {
	targetUsers, err := ol.GetUsers(ctx, utils.ListOptions{})
	if err != nil {
		return err
	}

	toAdd, toRemove, toUpdate, err := ol.BaseConfig.CompareUsers(source, targetUsers, ol.BaseConfig.Mapping["id"])
	if err != nil {
		return err
	}

	for _, u := range toAdd {
		mapped, err := ol.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		conv, err := MapToUser(mapped)
		if err != nil {
			return err
		}

		if _, err := ol.client.CreateUser(*conv); err != nil {
			return err
		}
	}

	for _, u := range toRemove {
		iID, err := strconv.Atoi(u[ol.BaseConfig.Mapping["id"]].(string))
		if err != nil {
			return err
		}

		if _, err := ol.client.DeleteUser(iID); err != nil {
			return err
		}
	}

	for _, u := range toUpdate {
		mapped, err := ol.BaseConfig.ConvertUserToProvider(u)
		if err != nil {
			return err
		}

		conv, err := MapToUser(mapped)
		if err != nil {
			return err
		}

		if _, err := ol.client.UpdateUser(int(conv.ID), *conv); err != nil {
			return err
		}
	}

	return nil
}

func MapToUser(user map[string]any) (*models.User, error) {
	bs, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	var original *models.User
	if err := json.Unmarshal(bs, &original); err != nil {
		return nil, err
	}

	return original, nil
}
