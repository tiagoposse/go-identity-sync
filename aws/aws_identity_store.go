package aws

import (
	"context"
	"fmt"
	"strings"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/tiagoposse/go-identity-sync/config"
	"github.com/tiagoposse/go-identity-sync/utils"
)

type awsIdentityStoreProvider struct {
	config.BaseConfig

	client          *identitystore.Client
	identityStoreID *string
}

func NewAwsIdentityStoreProvider(ctx context.Context, cfg *config.AwsIdentityStoreConfig) (*awsIdentityStoreProvider, error) {
	// Load AWS SDK configuration
	clicfg, err := awscfg.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS SDK configuration: %w", err)
	}

	// Create an identitystore client
	client := identitystore.NewFromConfig(clicfg)
	return &awsIdentityStoreProvider{
		BaseConfig:      cfg.BaseConfig,
		client:          client,
		identityStoreID: &cfg.IdentityStoreID,
	}, nil
}

func (aws *awsIdentityStoreProvider) GetUser(ctx context.Context, id string) {
	// Call GetUser API to get information for the specified user
	input := &identitystore.DescribeUserInput{
		IdentityStoreId: aws.identityStoreID,
		UserId:          utils.StrPtr(id),
	}

	output, err := aws.client.DescribeUser(context.Background(), input)
	if err != nil {
		fmt.Println("Error getting user information:", err)
		return
	}

	fmt.Printf("Username: %s\n", *output.UserName)
	fmt.Printf("User ID: %s\n", *output.UserId)
}

// func (aws *awsIdentityStoreProvider) GetUsers(ctx context.Context) ([]types.User, error) {
// 	return aws.SearchUsers(ctx, "")
// }

func (aws *awsIdentityStoreProvider) GetUsersAndMemberships(ctx context.Context, filter string) ([]types.User, map[string][]string, error) {
	input := &identitystore.ListUsersInput{}
	output, err := aws.client.ListUsers(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("listing users: %w", err)
	}

	groups := make(map[string][]string)

	filteredUsers := make([]types.User, 0)
	for _, user := range output.Users {
		if filter == "" ||
			strings.Contains(*user.UserName, filter) ||
			strings.Contains(*user.UserId, filter) ||
			strings.Contains(*user.DisplayName, filter) ||
			strings.Contains(*user.NickName, filter) {
			filteredUsers = append(filteredUsers, user)
		}
		userMemberships, err := aws.client.ListGroupMembershipsForMember(
			ctx,
			&identitystore.ListGroupMembershipsForMemberInput{},
		)
		if err != nil {
			return nil, nil, err
		}

		groups[*user.UserId] = make([]string, 0)
		for _, group := range userMemberships.GroupMemberships {
			groups[*user.UserId] = append(groups[*user.UserId], *group.GroupId)
		}
	}

	return filteredUsers, groups, nil
}

func (aws *awsIdentityStoreProvider) GetUsersConverted(ctx context.Context, filter string) ([]map[string]any, error) {
	users, memberships, err := aws.GetUsersAndMemberships(ctx, filter)
	if err != nil {
		return nil, err
	}

	convertedUsers := make([]map[string]any, 0)
	for _, item := range users {
		if user, err := aws.BaseConfig.ConvertUser(item); err != nil {
			return nil, err
		} else {
			user[aws.BaseConfig.GroupField] = memberships[item.Id]
			convertedUsers = append(convertedUsers, user)
		}
	}
	return convertedUsers, nil
}

func (aws *awsIdentityStoreProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := aws.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return aws.BaseConfig.CompareUsers(sourceUsers, users, aws.BaseConfig.Mapping["UserId"])
}
