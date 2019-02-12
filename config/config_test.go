package config

import (
	"os"
	"testing"
)

func TestNewAwsConfigFile(t *testing.T) {
	t.Run("ExplicitSource", func(t *testing.T) {
		f, err := NewAwsConfigFile(ConfFileName)
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()

		if f.Path != ConfFileName {
			t.Error("config file name mismatch")
			return
		}
	})

	t.Run("EnvVarSource", func(t *testing.T) {
		os.Setenv(ConfigFileEnvVar, ConfFileName)
		defer os.Unsetenv(ConfigFileEnvVar)

		f, err := NewAwsConfigFile(nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()

		if f.Path != ConfFileName {
			t.Error("config file name mismatch")
			return
		}
	})

	t.Run("BadSource", func(t *testing.T) {
		_, err := NewAwsConfigFile("not-my-file")
		if err == nil {
			t.Error("did not see expected error with bad file name")
			return
		}
	})
}

func TestAwsConfigFile_Profile(t *testing.T) {
	f, err := NewAwsConfigFile(ConfFileName)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	t.Run("Empty", func(t *testing.T) {
		s, err := f.Profile("")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != DefaultProfileName {
			t.Error("mismatched profile name")
			return
		}
	})

	t.Run("ConfigProfile", func(t *testing.T) {
		s, err := f.Profile("other")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != "profile other" {
			t.Error("mismatched profile name")
			return
		}
	})

	t.Run("NonConformingProfile", func(t *testing.T) {
		s, err := f.Profile("uncommon")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != "uncommon" {
			t.Error("mismatched profile name")
			return
		}
	})

	t.Run("BadProfile", func(t *testing.T) {
		_, err := f.Profile("not-good")
		if err == nil {
			t.Error("did not see expected error with bad profile name")
			return
		}
	})
}
