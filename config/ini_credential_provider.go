package config

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/service/iam"
	"os"
)

// CredentialsFileEnvVar is the credentials file environment variable name
const CredentialsFileEnvVar = "AWS_SHARED_CREDENTIALS_FILE"

// IniCredentialProvider enables the lookup of AWS credentials from an ini-formatted data source
type IniCredentialProvider struct {
	*awsConfigFile
}

// NewIniCredentialProvider initializes a default IniCredentialProvider using the specified source.  Valid sources
// include, a string representing a file path or url (file and http(s) urls supported), a Golang *url.URL, an []byte,
// a *os.File, or an io.Reader
func NewIniCredentialProvider(source interface{}) (*IniCredentialProvider, error) {
	cf, err := load(source, func(f *awsConfigFile) {
		s := defaults.SharedCredentialsFilename()
		if e, ok := os.LookupEnv(CredentialsFileEnvVar); ok {
			s = e
		}
		f.Path = s
		f.isTemp = false
	})
	if err != nil {
		return nil, err
	}

	return &IniCredentialProvider{cf}, nil
}

// Credentials will retrieve AWS credentials from the configured source location, for the provided profile.
// If the profile argument is nil or empty, the value of the AWS_PROFILE environment variable will be used, and if
// that isn't set, return credentials set in the "default" profile
func (p *IniCredentialProvider) Credentials(profile ...string) (credentials.Value, error) {
	v := credentials.Value{}

	if profile == nil || len(profile) < 1 {
		profile = []string{""}
	}

	pr, err := p.Profile(profile[0])
	if err != nil {
		return v, err
	}

	c := new(awsCredentials)
	if err := pr.MapTo(c); err != nil {
		return v, err
	}

	v.AccessKeyID = c.AccessKey
	v.SecretAccessKey = c.SecretKey
	v.SessionToken = c.SessionToken
	if !v.HasKeys() {
		return v, fmt.Errorf("incomplete credentials, missing access key and/or secret key")
	}

	return v, nil
}

// UpdateCredentials updates the given profile with the provided credentials.  The creds can be an iam.AccessKey or
// credentials.Value type (or pointers to either).  Updates are only made to the in-memory representation of the data,
// it is the caller's responsibility to persist the information to storage, either via the SaveTo() or WriteTo() methods.
func (p *IniCredentialProvider) UpdateCredentials(profile string, creds interface{}) error {
	c := new(awsCredentials)

	switch t := creds.(type) {
	case nil:
		return nil
	case iam.AccessKey:
		if *t.Status == iam.StatusTypeActive {
			c.AccessKey = *t.AccessKeyId
			c.SecretKey = *t.SecretAccessKey
		} else {
			return nil
		}
	case *iam.AccessKey:
		if *t.Status == iam.StatusTypeActive {
			c.AccessKey = *t.AccessKeyId
			c.SecretKey = *t.SecretAccessKey
		} else {
			return nil
		}
	case credentials.Value:
		c.AccessKey = t.AccessKeyID
		c.SecretKey = t.SecretAccessKey
		c.SessionToken = t.SessionToken
	case *credentials.Value:
		c.AccessKey = t.AccessKeyID
		c.SecretKey = t.SecretAccessKey
		c.SessionToken = t.SessionToken
	default:
		return fmt.Errorf("unsupported credential type")
	}

	s, err := p.Profile(profile)
	if err != nil {
		return err
	}

	return s.ReflectFrom(c)
}
