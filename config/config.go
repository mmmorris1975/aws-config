package config

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/go-ini/ini"
	"os"
)

// ConfigFileEnvVar is the configuration file environment variable name
const ConfigFileEnvVar = "AWS_CONFIG_FILE"

// AwsConfigFile is the object used to access profile data in the AWS SDK config file
type AwsConfigFile struct {
	*awsConfigFile
}

// NewAwsConfigFile creates a new AwsConfigFile object from the provides source
func NewAwsConfigFile(source interface{}) (*AwsConfigFile, error) {
	c, err := load(source, func(f *awsConfigFile) {
		s := defaults.SharedConfigFilename()
		if e, ok := os.LookupEnv(ConfigFileEnvVar); ok {
			s = e
		}
		f.path = s
		f.isTemp = false
	})
	if err != nil {
		return nil, err
	}

	return &AwsConfigFile{c}, nil
}

// Profile overrides the default Profile lookup logic to include a callback to re-try the lookup with "profile"
// appended to the name
func (f *AwsConfigFile) Profile(profile string) (*ini.Section, error) {
	return f.profile(profile, func(n string) string {
		return fmt.Sprintf("profile %s", n)
	})
}
