package config

import (
	"encoding/json"
	"fmt"

	resolvers "github.com/tiagoposse/go-secret-resolvers"
)

type Config struct {
	Okta             *OktaConfig             `yaml:"okta"`
	Google           *GoogleConfig           `yaml:"google"`
	Gitlab           *GitlabConfig           `yaml:"gitlab"`
	Github           *GithubConfig           `yaml:"github"`
	AwsIAM           *AwsIAMConfig           `yaml:"awsIAM"`
	AwsIdentityStore *AwsIdentityStoreConfig `yaml:"awsIdentityStore"`
	Gcp              *GcpConfig              `yaml:"gcp"`
	Azure            *AzureConfig            `yaml:"azure"`
	Ldap             *LdapConfig             `yaml:"ldap"`
	ActiveDirectory  *ADConfig               `yaml:"ad"`
	OneLogin         *OneLoginConfig         `yaml:"onelogin"`
	Keycloak         *KeycloakConfig         `yaml:"keycloak"`
}

type OktaConfig struct {
	BaseConfig `yaml:",inline"`
	Domain     string                   `yaml:"domain"`
	Token      *resolvers.ResolverField `yaml:"token"`
}

type GoogleConfig struct {
	BaseConfig        `yaml:",inline"`
	Domain            string                   `yaml:"domain"`
	ServiceAccountKey *resolvers.ResolverField `yaml:"serviceAccountKey"`
	UserToImpersonate string                   `yaml:"userToImpersonate"`
}

type GitlabConfig struct {
	BaseConfig   `yaml:",inline"`
	Url          string                   `yaml:"url"`
	Organisation string                   `yaml:"org"`
	Token        *resolvers.ResolverField `yaml:"token"`
}

type GithubConfig struct {
	BaseConfig   `yaml:",inline"`
	Organisation string                  `yaml:"org"`
	Token        resolvers.ResolverField `yaml:"token"`
}

type AwsIAMConfig struct {
	BaseConfig `yaml:",inline"`
	Profile    *string `yaml:"profile"`
}

type AwsIdentityStoreConfig struct {
	BaseConfig      `yaml:",inline"`
	IdentityStoreID string  `yaml:"storeID"`
	Profile         *string `yaml:"profile"`
}

type AzureConfig struct {
	BaseConfig `yaml:",inline"`
}

type GcpConfig struct {
	BaseConfig `yaml:",inline"`
}

type LdapConfig struct {
	BaseConfig `yaml:",inline"`
}

type OneLoginConfig struct {
	ClientID     *resolvers.ResolverField `yaml:"clientID"`
	ClientSecret *resolvers.ResolverField `yaml:"clientSecret"`
	BaseConfig   `yaml:",inline"`
}

type KeycloakConfig struct {
	BaseConfig   `yaml:",inline"`
	Url          string                   `yaml:"string"`
	Realm        string                   `yaml:"realm"`
	ClientID     *resolvers.ResolverField `yaml:"clientID"`
	ClientSecret *resolvers.ResolverField `yaml:"clientSecret"`
	Username     *resolvers.ResolverField `yaml:"username"`
	Password     *resolvers.ResolverField `yaml:"password"`
}

type ADConfig struct {
	BaseConfig `yaml:",inline"`
}

type BaseConfig struct {
	IgnoreUsers  []string          `yaml:"ignoreUsers"`
	IgnoreGroups []string          `yaml:"ignoreGroups"`
	GroupFilters []string          `yaml:"groupFilters"`
	UserFilters  []string          `yaml:"userFilters"`
	Mapping      map[string]string `yaml:"mapping"`
}

func (bc BaseConfig) ConvertUsers(arr any) ([]map[string]any, error) {
	sourceUsers := make([]map[string]any, 0)
	for _, item := range arr.([]any) {
		if user, err := bc.ConvertUser(item); err != nil {
			return nil, err
		} else {
			sourceUsers = append(sourceUsers, user)
		}
	}

	return sourceUsers, nil
}

func (bc BaseConfig) ConvertUser(user any) (map[string]any, error) {
	bs, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	var original map[string]any
	if err := json.Unmarshal(bs, &original); err != nil {
		return nil, err
	}

	converted := make(map[string]any)
	for k, v := range bc.Mapping {
		if val, ok := original[k]; !ok {
			return nil, fmt.Errorf("field %s does not exist for user: %v", k, original)
		} else {
			converted[v] = val
		}
	}

	return converted, nil
}

func (bc BaseConfig) CompareUsers(source any, target []map[string]any, field string) (toAdd, toRemove, toUpdate []map[string]any, retErr error) {
	sourceMap := make(map[string]any)

	// Create a map from source array for easier comparison
	for _, item := range source.([]any) {
		if user, err := bc.ConvertUser(item); err != nil {
			retErr = err
			return
		} else {
			sourceMap[user[field].(string)] = item
		}
	}

	// Compare each element in the target array
	for _, item := range target {
		key := item[field].(string)

		if sourceItem, ok := sourceMap[key]; !ok {
			// Item in target not found in source, add toAdd
			toAdd = append(toAdd, item)
		} else if !compareMaps(sourceItem.(map[string]any), item) {
			// Item found but content is different, add toUpdate
			toUpdate = append(toUpdate, item)
		}

		// Remove the key from source map to identify items that need to be removed from target later
		delete(sourceMap, key)
	}

	// Any remaining items in sourceMap need to be removed from target
	for _, value := range sourceMap {
		toRemove = append(toRemove, value.(map[string]any))
	}

	return toAdd, toRemove, toUpdate, nil
}

// Function to compare two maps
func compareMaps(map1, map2 map[string]any) bool {
	// Compare maps based on your desired criteria
	// For simplicity, this example assumes that the maps are equal if all keys and values match
	for key, value := range map1 {
		if map2Val, ok := map2[key]; !ok || map2Val != value {
			return false
		}
	}
	return true
}
