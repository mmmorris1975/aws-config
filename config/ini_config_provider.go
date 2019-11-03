package config

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/go-ini/ini"
	"os"
	"sort"
	"strings"
)

// ConfigFileEnvVar is the configuration file environment variable name
const ConfigFileEnvVar = "AWS_CONFIG_FILE"

// IniConfigProvider enables the lookup of AWS configuration from an ini-formatted data source
type IniConfigProvider struct {
	*awsConfigFile
}

// NewIniConfigProvider initializes a default IniConfigProvider using the specified source.  Valid sources
// include, a string representing a file path or url (file and http(s) urls supported), a Golang *url.URL, an []byte,
// a *os.File, or an io.Reader
func NewIniConfigProvider(source interface{}) (*IniConfigProvider, error) {
	cf, err := load(source, func(f *awsConfigFile) {
		s := defaults.SharedConfigFilename()
		if e, ok := os.LookupEnv(ConfigFileEnvVar); ok {
			s = e
		}
		f.Path = s
		f.isTemp = false
	})
	if err != nil {
		return nil, err
	}

	return &IniConfigProvider{cf}, nil
}

// Config will return the configuration attributes for the specified profile.  If the profile is nil, the
// configuration of the default profile will be returned.
func (p *IniConfigProvider) Config(profile ...string) (*AwsConfig, error) {
	c := new(AwsConfig)

	if profile == nil || len(profile) < 1 {
		profile = []string{DefaultProfileName}
	}

	s, err := p.Profile(profile[0])
	if err != nil {
		return nil, err
	}

	if err := s.MapTo(c); err != nil {
		return nil, err
	}

	c.rawAttributes = s.KeysHash()
	c.Profile = profile[0]

	return c, nil
}

// Profile overrides the default Profile lookup logic to include a callback
// to re-try the lookup with "profile " prepended to the name
func (p *IniConfigProvider) Profile(profile string) (*ini.Section, error) {
	return p.profile(profile, func(n string) string {
		return fmt.Sprintf("profile %s", n)
	})
}

// ListProfiles will return an array of profile names found in the config file.  If the roles arg is false,
// then all profile sections found in the config file will be returned; otherwise only profile sections which
// have the role_arn property will be returned.
func (p *IniConfigProvider) ListProfiles(roles bool) []string {
	profiles := make([]string, 0)

	for _, s := range p.Sections() {
		if s.Name() == ini.DefaultSection {
			continue
		}

		n := strings.TrimPrefix(s.Name(), "profile ")
		if roles {
			if s.HasKey("role_arn") {
				profiles = append(profiles, n)
			}
		} else {
			profiles = append(profiles, n)
		}
	}

	sort.Strings(profiles)
	return profiles
}
