package config

import "strconv"

type awsConfigResolver struct {
	lookupDefaultProfile bool
	lookupSourceProfile  bool
	configProvider       AwsConfigProvider
}

// Create a default AWS config resolver which will lookup information in the INI config source for the default profile,
// and source_profile (if configured), and merge with the data in the provided profile
func NewAwsConfigResolver(source interface{}) (*awsConfigResolver, error) {
	cp, err := NewIniConfigProvider(source)
	if err != nil {
		return nil, err
	}

	return &awsConfigResolver{
		lookupDefaultProfile: true,
		lookupSourceProfile:  true,
		configProvider:       cp,
	}, nil
}

// LookupDefaultProfile is a fluent method for enabling (or disabling) the inclusion of default profile
// data in the resolved configuration
func (r *awsConfigResolver) WithLookupDefaultProfile(b bool) *awsConfigResolver {
	r.lookupDefaultProfile = b
	return r
}

// LookupSourceProfile is a fluent method for enabling (or disabling) the inclusion of source_profile
// data in the resolved configuration, if the source_profile attribute is found in the target profile
func (r *awsConfigResolver) WithLookupSourceProfile(b bool) *awsConfigResolver {
	r.lookupSourceProfile = b
	return r
}

// WithConfigProvider is a fluent method for setting the AwsConfigProvider type of the resolver.
// This will be used to do the work in the Resolve() method
func (r *awsConfigResolver) WithConfigProvider(p AwsConfigProvider) *awsConfigResolver {
	r.configProvider = p
	return r
}

// Merge will combine the attributes of the provided AwsConfig types and return it as a single AwsConfig.
// Objects later in the input list will overwrite values in earlier objects if the value for the attribute is
// not empty, or the explict string "0"
func (r *awsConfigResolver) Merge(config ...*AwsConfig) (*AwsConfig, error) {
	c := new(AwsConfig)
	c.rawAttributes = make(map[string]string)

	for _, x := range config {
		for k, v := range x.rawAttributes {
			if len(v) > 0 && v != "0" {
				c.rawAttributes[k] = v
			}
		}

		if len(x.Profile) > 0 {
			c.Profile = x.Profile
		}
	}

	if len(c.rawAttributes) > 0 {
		durSec, err := strconv.ParseInt(c.Get("duration_seconds"), 0, 0)
		if err != nil {
			durSec = 0
		}

		c.CaBundle = c.Get("ca_bundle")
		c.CredentialSource = c.Get("credential_source")
		c.DurationSeconds = int(durSec)
		c.ExternalId = c.Get("external_id")
		c.MfaSerial = c.Get("mfa_serial")
		c.Region = c.Get("region")
		c.RoleArn = c.Get("role_arn")
		c.RoleSessionName = c.Get("role_session_name")
		c.SourceProfile = c.Get("source_profile")
	}

	return c, nil
}

func (r *awsConfigResolver) Resolve(profile ...string) (*AwsConfig, error) {
	if profile == nil || len(profile) < 1 {
		// quick path ... return default profile data
		return r.configProvider.Config()
	}

	c := make([]*AwsConfig, 0)

	p, err := r.configProvider.Config(profile...)
	if err != nil {
		return nil, err
	}

	if r.lookupDefaultProfile {
		d, err := r.configProvider.Config()
		if err != nil {
			return nil, err
		}
		c = append(c, d)
	}

	if r.lookupSourceProfile && len(p.SourceProfile) > 0 {
		s, err := r.configProvider.Config(p.SourceProfile)
		if err != nil {
			return nil, err
		}
		c = append(c, s)
	}

	return r.Merge(append(c, p)...)
}
