package config

import (
	"os"
	"testing"
)

func TestNewIniConfigProvider(t *testing.T) {
	t.Run("explicit source", func(t *testing.T) {
		f, err := NewIniConfigProvider(ConfFileName)
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()

		if f.Path != ConfFileName {
			t.Error("path mis-match")
		}
	})

	t.Run("env var source", func(t *testing.T) {
		os.Setenv(ConfigFileEnvVar, ConfFileName)
		defer os.Unsetenv(ConfigFileEnvVar)

		f, err := NewIniConfigProvider(nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()

		if f.Path != ConfFileName {
			t.Error("path mis-match")
		}
	})

	t.Run("bad source", func(t *testing.T) {
		_, err := NewIniConfigProvider("not-my-file")
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestIniConfigProvider_Profile(t *testing.T) {
	f, err := NewIniConfigProvider(ConfFileName)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	// Should return default profile
	t.Run("empty", func(t *testing.T) {
		s, err := f.Profile("")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != DefaultProfileName {
			t.Error("mismatched profile name")
		}
	})

	t.Run("aws format", func(t *testing.T) {
		s, err := f.Profile("other")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != "profile other" {
			t.Error("mismatched profile name")
		}
	})

	t.Run("not aws format", func(t *testing.T) {
		s, err := f.Profile("uncommon")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != "uncommon" {
			t.Error("mismatched profile name")
		}
	})

	t.Run("bad profile name", func(t *testing.T) {
		_, err := f.Profile("not-a-profile")
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestIniConfigProvider_Config(t *testing.T) {
	f, err := NewIniConfigProvider(ConfFileName)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	// Should return default profile
	t.Run("empty", func(t *testing.T) {
		c, err := f.Config()
		if err != nil {
			t.Error(err)
			return
		}
		t.Logf("%+v", c)
		if c.Profile != DefaultProfileName || c.Region != "us-east-2" {
			t.Error("data mismatch")
		}
	})

	t.Run("aws format", func(t *testing.T) {
		c, err := f.Config("other")
		if err != nil {
			t.Error(err)
			return
		}

		if c.Profile != "other" || c.Region != "us-west-1" {
			t.Error("data mismatch")
		}
	})

	t.Run("not aws format", func(t *testing.T) {
		c, err := f.Config("uncommon")
		if err != nil {
			t.Error(err)
			return
		}

		if c.Profile != "uncommon" || c.Region != "eu-west-1" {
			t.Error("data mismatch")
		}
	})

	t.Run("bad profile name", func(t *testing.T) {
		_, err := f.Config("not-a-profile")
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestIniConfigProvider_ListProfiles(t *testing.T) {
	f, err := NewIniConfigProvider(ConfFileName)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

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
