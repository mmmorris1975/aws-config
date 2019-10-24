package config

import "github.com/aws/aws-sdk-go/aws/credentials"

// AwsConfig is the type containing the explicitly supported AWS SDK configuration attributes
type AwsConfig struct {
	CaBundle         string `ini:"ca_bundle" env:"AWS_CA_BUNDLE"`
	CredentialSource string `ini:"credential_source"`
	DurationSeconds  int    `ini:"duration_seconds" env:"DURATION_SECONDS,CREDENTIALS_DURATION"`
	ExternalId       string `ini:"external_id" env:"EXTERNAL_ID"`
	MfaSerial        string `ini:"mfa_serial" env:"MFA_SERIAL"`
	Profile          string `env:"AWS_PROFILE"`
	Region           string `ini:"region" env:"AWS_REGION,AWS_DEFAULT_REGION"`
	RoleArn          string `ini:"role_arn"`
	RoleSessionName  string `ini:"role_session_name" env:"AWS_ROLE_SESSION_NAME"`
	SourceProfile    string `ini:"source_profile"`
	rawAttributes    map[string]string
}

type awsCredentials struct {
	AccessKey    string `ini:"aws_access_key_id" env:"AWS_ACCESS_KEY_ID,AWS_ACCESS_KEY"`
	SecretKey    string `ini:"aws_secret_access_key" env:"AWS_SECRET_ACCESS_KEY,AWS_SECRET_KEY"`
	SessionToken string `ini:"aws_session_token" env:"AWS_SESSION_TOKEN,AWS_SECURITY_TOKEN"`
}

// AwsConfigProvider is an interface defining the contract for conforming types to provide AWS configuration
type AwsConfigProvider interface {
	Config(profile ...string) (*AwsConfig, error)
}

// AwsCredentialProvider is an interface defining the contract for conforming types to provide AWS credentials
type AwsCredentialProvider interface {
	Credentials(profile ...string) (credentials.Value, error)
}

type AwsConfigResolver interface {
	Merge(config ...*AwsConfig) (*AwsConfig, error)
	Resolve(profile ...string) (*AwsConfig, error)
}
