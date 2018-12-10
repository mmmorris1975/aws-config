package credentials

import (
	. "aws-config"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/go-ini/ini"
	"os"
)

// credential file env var name
const CredFileEnvVar = "AWS_SHARED_CREDENTIALS_FILE"

type awsCredentials struct {
	AccessKey string `ini:"aws_access_key_id"`
	SecretKey string `ini:"aws_secret_access_key"`
}

type AwsCredentialsFile struct {
	*ini.File
	Path string
}

// load the file from the provides source, which may be a string representing a file name,
// or an []byte of raw data.  If source is nil, it will check the env var AWS_SHARED_CREDENTIALS_FILE;
// if the env var is not set, it will use the SDK default credential file name.
func Load(source interface{}) (*AwsCredentialsFile, error) {
	if source == nil {
		if v, ok := os.LookupEnv(CredFileEnvVar); ok {
			source = v
		} else {
			source = defaults.SharedCredentialsFilename()
		}
	}

	s, err := ini.Load(source)
	if err != nil {
		return nil, err
	}
	s.BlockMode = true

	c := &AwsCredentialsFile{File: s}
	switch t := source.(type) {
	case string:
		c.Path = t
	default:
		c.Path = ""
	}

	return c, nil
}

// Retrieve the credentials for a given profile name, and provide them as a credentials.Value type
// Returns an error if the aws_access_key_id and/or aws_secret_access_key properties are missing/unset
func (f *AwsCredentialsFile) Credentials(profile string) (credentials.Value, error) {
	p := ResolveProfile(&profile)
	s, err := f.GetSection(p)
	if err != nil {
		return credentials.Value{}, err
	}

	c := new(awsCredentials)
	if err := s.MapTo(c); err != nil {
		return credentials.Value{}, err
	}

	if len(c.AccessKey) < 1 || len(c.SecretKey) < 1 {
		return credentials.Value{}, fmt.Errorf("incomplete credentials, missing access key and/or secret key")
	}

	return credentials.Value{AccessKeyID: c.AccessKey, SecretAccessKey: c.SecretKey}, nil
}

// Update the credentials for a given profile, with the provided credentials.  The creds can be
// an iam.AccessKey or credentials.Value type (or pointers to either).  Updates are only made to
// the in-memory representation of the data, it is the caller's responsibility to persist the
// information to storage, either via the SaveTo() or WriteTo() methods.
func (f *AwsCredentialsFile) UpdateCredentials(profile string, creds interface{}) error {
	// this feels sort of dangerous, but does allow us to obey env vars
	p := ResolveProfile(&profile)

	var c awsCredentials
	switch t := creds.(type) {
	case nil:
		return fmt.Errorf("nil credentials provided")
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
		// smooth one AWS!  the credentials.Value and iam.AccessKey fields are just slliiiiightly different
		c.AccessKey = t.AccessKeyID
		c.SecretKey = t.SecretAccessKey
	case *credentials.Value:
		c.AccessKey = t.AccessKeyID
		c.SecretKey = t.SecretAccessKey
	default:
		return fmt.Errorf("unsupported credential type: %v", t)
	}

	s, err := f.NewSection(p)
	if err != nil {
		return err
	}

	if err := s.ReflectFrom(&c); err != nil {
		return err
	}

	return nil
}
