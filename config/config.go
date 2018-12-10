package config

import (
	. "aws-config"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/go-ini/ini"
	"os"
)

// config file env var name
const ConfFileEnvVar = "AWS_CONFIG_FILE"

type AwsConfigFile struct {
	*ini.File
	Path string
}

// load the file from the provides source, which may be a string representing a file name,
// or an []byte of raw data.  If source is nil, it will check the env var AWS_CONFIG_FILE;
// if the env var is not set, it will use the SDK default credential file name.
func Load(source interface{}) (*AwsConfigFile, error) {
	if source == nil {
		if v, ok := os.LookupEnv(ConfFileEnvVar); ok {
			source = v
		} else {
			source = defaults.SharedConfigFilename()
		}
	}

	s, err := ini.Load(source)
	if err != nil {
		return nil, err
	}
	s.BlockMode = true

	c := &AwsConfigFile{File: s}
	switch t := source.(type) {
	case string:
		c.Path = t
	default:
		c.Path = ""
	}

	return c, nil
}

// Retrieve the configuration for the provided profile.  Since there is no explicit format of the profile
// data (the cli/sdk defines some values, but allows custom attributes), the returned value is the ini file
// section data, where the call will be able to process the profile attributes locally.
func (f *AwsConfigFile) Profile(profile string) (*ini.Section, error) {
	var s *ini.Section
	p := ResolveProfile(&profile)

	s, err := f.GetSection(p)
	if err != nil {
		// The AWS cli/sdk config file has this strange format where non-default profile names are prefixed
		// with the word 'profile', however that same convention does not apply to the credentials file.
		if p == DefaultProfileName {
			return nil, err
		}

		return f.GetSection(fmt.Sprintf("profile %s", p))
	}

	return s, nil
}
