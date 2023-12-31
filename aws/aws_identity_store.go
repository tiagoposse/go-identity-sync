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

	// Print information about the user
	fmt.Printf("Username: %s\n", *output.UserName)
	fmt.Printf("User ID: %s\n", *output.UserId)
	// Add more fields as needed
}

func (aws *awsIdentityStoreProvider) GetUsers(ctx context.Context) ([]types.User, error) {
	return aws.SearchUsers(ctx, "")
}

func (aws *awsIdentityStoreProvider) SearchUsers(ctx context.Context, filter string) ([]types.User, error) {
	input := &identitystore.ListUsersInput{}
	output, err := aws.client.ListUsers(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	filteredUsers := make([]types.User, 0)
	for _, user := range output.Users {
		if filter == "" ||
			strings.Contains(*user.UserName, filter) ||
			strings.Contains(*user.UserId, filter) ||
			strings.Contains(*user.DisplayName, filter) ||
			strings.Contains(*user.NickName, filter) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	return filteredUsers, nil
}

func (aws *awsIdentityStoreProvider) Sync(ctx context.Context, users []map[string]any) (add, remove, update []map[string]any, retErr error) {
	sourceUsers, err := aws.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return aws.BaseConfig.CompareUsers(sourceUsers, users, aws.BaseConfig.Mapping["UserId"])
}
