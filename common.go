package aws_config

import (
	"github.com/go-ini/ini"
	"os"
	"strings"
)

const (
	ProfileEnvVar        = "AWS_PROFILE"
	DefaultProfileEnvVar = "AWS_DEFAULT_PROFILE"
)

var (
	DefaultProfileName = strings.ToLower(ini.DEFAULT_SECTION)
)

// Resolve the provided profile name.  If the passed value is nil or empty, check the AWS_PROFILE and
// AWS_DEFAULT_PROFILE env vars (in that order).  If nothings is found, return "default"
func ResolveProfile(p *string) string {
	if p == nil || len(*p) < 1 {
		if v, ok := os.LookupEnv(ProfileEnvVar); ok {
			p = &v
		} else if v, ok := os.LookupEnv(DefaultProfileEnvVar); ok {
			p = &v
		} else {
			p = &DefaultProfileName
		}
	}

	return *p
}
