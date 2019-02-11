package config

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/service/iam"
	"os"
)

const CredentialsFileEnvVar = "AWS_SHARED_CREDENTIALS_FILE"

type awsCredentials struct {
	AccessKey string `ini:"aws_access_key_id"`
	SecretKey string `ini:"aws_secret_access_key"`
}

type AwsCredentialsFile struct {
	*awsConfigFile
}

func NewAwsCredentialsFile(source interface{}) (*AwsCredentialsFile, error) {
	c, err := load(source, func(f *awsConfigFile) {
		s := defaults.SharedCredentialsFilename()
		if e, ok := os.LookupEnv(CredentialsFileEnvVar); ok {
			s = e
		}
		f.path = s
		f.isTemp = false
	})
	if err != nil {
		return nil, err
	}

	return &AwsCredentialsFile{c}, nil
}

// Retrieve the credentials for a given profile name, and provide them as a credentials.Value type
// Returns an error if the aws_access_key_id and/or aws_secret_access_key properties are missing/unset
func (f *AwsCredentialsFile) Credentials(profile string) (credentials.Value, error) {
	s, err := f.Profile(profile)
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

	s, err := f.Profile(profile)
	if err != nil {
		return err
	}

	if err := s.ReflectFrom(&c); err != nil {
		return err
	}

	return nil
}
