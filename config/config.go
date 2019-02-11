package config

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/go-ini/ini"
	"os"
)

const ConfigFileEnvVar = "AWS_CONFIG_FILE"

type AwsConfigFile struct {
	*awsConfigFile
}

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

// Override the default Profile lookup logic to include a callback to re-try the lookup with "profile" appended to the name
func (f *AwsConfigFile) Profile(profile string) (*ini.Section, error) {
	return f.profile(profile, func(n string) string {
		return fmt.Sprintf("profile %s", n)
	})
}
