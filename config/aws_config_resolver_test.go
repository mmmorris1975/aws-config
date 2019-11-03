package config

import "testing"

func TestNewAwsConfigResolver(t *testing.T) {
	t.Run("good source", func(t *testing.T) {
		_, err := NewAwsConfigResolver(ConfFileName)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("nil", func(t *testing.T) {
		_, err := NewAwsConfigResolver(nil)
		if err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("bad source", func(t *testing.T) {
		_, err := NewAwsConfigResolver("not-a-source")
		if err == nil {
			t.Error("did not receive expected error")
		}
	})
}

func TestAwsConfigResolver_Fluent(t *testing.T) {
	r := new(awsConfigResolver).
		WithConfigProvider(NewEnvConfigProvider()).
		WithLookupDefaultProfile(false).
		WithLookupSourceProfile(false)

	if r.configProvider == nil || r.lookupDefaultProfile || r.lookupSourceProfile {
		t.Error("config mismatch")
	}
}

func TestAwsConfigResolver_Resolve(t *testing.T) {
	r, err := NewAwsConfigResolver(ConfFileName)
	if err != nil {
		t.Error(err)
		return
	}

	// should return the default profile data
	t.Run("nil profile", func(t *testing.T) {
		c, err := r.Resolve()
		if err != nil {
			t.Error(err)
			return
		}

		if c.Profile != DefaultProfileName || c.Region != "us-east-2" {
			t.Error("data mismatch")
		}
	})

	t.Run("simple profile", func(t *testing.T) {
		c, err := r.Resolve("other")
		if err != nil {
			t.Error(err)
			return
		}

		if c.Profile != "other" || c.Region != "us-west-1" || len(c.Get("custom_attribute")) < 1 {
			t.Error("data mismatch")
		}
	})

	t.Run("mfa profile", func(t *testing.T) {
		c, err := r.Resolve("mfa")
		if err != nil {
			t.Error(err)
			return
		}

		if c.Profile != "mfa" || c.Region != "ap-southeast-2" || c.SourceProfile != DefaultProfileName ||
			c.ExternalId != "qq" || len(c.MfaSerial) < 1 || len(c.RoleArn) < 1 || len(c.rawAttributes) < 5 {
			t.Error("data mismatch")
		}
	})

	t.Run("bad profile", func(t *testing.T) {
		_, err := r.Resolve("not-a-profile")
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestAwsConfigResolver_ListProfiles(t *testing.T) {
	f, err := NewAwsConfigResolver(ConfFileName)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("arg true", func(t *testing.T) {
		profiles := f.ListProfiles(true)
		if len(profiles) != 1 {
			t.Error("did not find expected number of role profile sections")
		}
	})

	t.Run("arg false", func(t *testing.T) {
		profiles := f.ListProfiles(false)
		if len(profiles) < 5 {
			t.Error("did not find expected number of role profile sections")
		}
	})
}
