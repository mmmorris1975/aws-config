package config

type awsConfigResolver struct {
	lookupDefaultProfile bool
	lookupSourceProfile  bool
}

// Create a default AWS config resolver which will lookup information in the default profile,
// and source_profile (if specified), and merge with the data in the provided profile
func NewAwsConfigResolver() *awsConfigResolver {
	return &awsConfigResolver{
		lookupDefaultProfile: true,
		lookupSourceProfile:  true,
	}
}

// LookupDefaultProfile is a fluent method for enabling (or disabling) the inclusion of default profile
// data in the resolved configuration
func (r *awsConfigResolver) LookupDefaultProfile(b bool) *awsConfigResolver {
	r.lookupDefaultProfile = b
	return r
}

// LookupSourceProfile is a fluent method for enabling (or disabling) the inclusion of source_profile
// data in the resolved configuration, if the source_profile attribute is found in the target profile
func (r *awsConfigResolver) LookupSourceProfile(b bool) *awsConfigResolver {
	r.lookupSourceProfile = b
	return r
}

func (r *awsConfigResolver) Merge(config ...*AwsConfig) (*AwsConfig, error) {
	return &AwsConfig{}, nil
}

func (r *awsConfigResolver) Resolve(profile ...string) (*AwsConfig, error) {
	return &AwsConfig{}, nil
}
