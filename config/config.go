package config

import (
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
