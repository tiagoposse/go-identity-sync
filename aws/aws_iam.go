package aws

import (
	"context"
	"fmt"
	"strings"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/tiagoposse/go-identity-sync/config"
	"github.com/tiagoposse/go-identity-sync/utils"
)

type awsIAMProvider struct {
	config.BaseConfig

	client *iam.Client
}

func NewAwsIAMProvider(ctx context.Context, cfg *config.AwsIAMConfig) (*awsIAMProvider, error) {
	// Load AWS SDK configuration
	clicfg, err := awscfg.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("loading AWS SDK configuration: %w", err)
	}

	// Create an IAM client
	client := iam.NewFromConfig(clicfg)
	return &awsIAMProvider{
		BaseConfig: cfg.BaseConfig,
		client:     client,
	}, nil
}

func (aws *awsIAMProvider) GetUser(ctx context.Context, id string) (*types.User, error) {
	input := &iam.GetUserInput{
		UserName: utils.StrPtr(id),
	}

	output, err := aws.client.GetUser(context.Background(), input)
	if err != nil {
		return nil, fmt.Errorf("getting user information: %w", err)
	}

	return output.User, nil
}

func (aws *awsIAMProvider) GetUsers(ctx context.Context) ([]types.User, error) {
	return aws.SearchUsers(ctx, "")
}

func (aws *awsIAMProvider) SearchUsers(ctx context.Context, filter string) ([]types.User, error) {
	input := &iam.ListUsersInput{}
	output, err := aws.client.ListUsers(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("listing users: %w", err)
	}

	filteredUsers := make([]types.User, 0)
	for _, user := range output.Users {
		if filter == "" || strings.Contains(*user.UserName, filter) || strings.Contains(*user.UserId, filter) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	return filteredUsers, nil
}

func (aws *awsIAMProvider) Sync(ctx context.Context, target []map[string]any) (add, remove, update []map[string]any, retErr error) {
	currentUsers, err := aws.GetUsers(ctx)
	if err != nil {
		retErr = err
		return
	}

	return aws.BaseConfig.CompareUsers(currentUsers, target, aws.BaseConfig.Mapping["UserId"])
}
