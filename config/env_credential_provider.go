package config

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
)

// EnvCredentialProfile enables the lookup of AWS credentials from environment variables
type EnvCredentialProvider uint8

// NewEnvCredentialProvider initializes a default EnvCredentialProvider
func NewEnvCredentialProvider() *EnvCredentialProvider {
	return new(EnvCredentialProvider)
}

// Credentials will retrieve AWS credentials from the SDK supported environment variables.
// For Access Keys, these are ... AWS_ACCESS_KEY_ID and AWS_ACCESS_KEY
// for Secret Keys, these are ... AWS_SECRET_ACCESS_KEY and AWS_SECRET_KEY
// for Session Tokens, this is ... AWS_SESSION_TOKEN
func (p *EnvCredentialProvider) Credentials(profile ...string) (credentials.Value, error) {
	ec := credentials.NewEnvCredentials()
	return ec.Get()
}
