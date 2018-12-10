package config

import (
	. "aws-config"
	"os"
	"testing"
)

const ConfFile = "test-conf"

func TestLoad(t *testing.T) {
	t.Run("valid-string", func(t *testing.T) {
		if _, err := Load(ConfFile); err != nil {
			t.Error("failed to load config file")
			return
		}
	})

	t.Run("invalid-string", func(t *testing.T) {
		if _, err := Load("invalid"); err == nil {
			t.Error("unexpectedly loaded an invalid file location")
			return
		}
	})

	t.Run("bytes-valid", func(t *testing.T) {
		v := []byte("[default]")
		if _, err := Load(v); err != nil {
			t.Errorf("failed to load []byte data: %v", err)
			return
		}
	})

	t.Run("bytes-invalid", func(t *testing.T) {
		v := []byte("data")
		if _, err := Load(v); err == nil {
			t.Error("unexpectedly loaded invalid []byte data")
			return
		}
	})

	t.Run("nil-envvar", func(t *testing.T) {
		os.Setenv(ConfFileEnvVar, ConfFile)
		defer os.Unsetenv(ConfFileEnvVar)

		if _, err := Load(nil); err != nil {
			t.Errorf("failed to load from env var: %v", err)
			return
		}
	})

	t.Run("nil-default", func(t *testing.T) {
		// this one will be tricky, since it will check the default location, which may
		// or may not exist on the local system.  Best we can probably test for is no panics
		defer func() {
			if r := recover(); r != nil {
				t.Error("panic when trying to load default conf file")
				return
			}
		}()

		Load(nil)
	})
}

func TestAwsConfigFile_Profile(t *testing.T) {
	f, err := Load(ConfFile)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("explicit-default", func(t *testing.T) {
		if _, err := f.Profile("default"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("empty", func(t *testing.T) {
		if _, err := f.Profile(""); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("non-default", func(t *testing.T) {
		s, err := f.Profile("other")
		if err != nil {
			t.Error(err)
			return
		}

		if _, err := s.GetKey("aws_api_key_duration"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("missing-section", func(t *testing.T) {
		if _, err := f.Profile("missing"); err == nil {
			t.Error("succeeded in loading a missing profile")
			return
		}
	})

	t.Run("empty-section", func(t *testing.T) {
		if _, err := f.Profile("empty"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("bare-profile", func(t *testing.T) {
		// While loading a non-default profile which isn't prefixed with 'profile' is discouraged by the AWS
		// cli/sdk, it's not bad form for the ini-file format.
		if _, err := f.Profile("bad-ish"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("profile-env-var", func(t *testing.T) {
		os.Setenv(ProfileEnvVar, "other")
		defer os.Unsetenv(ProfileEnvVar)

		s, err := f.Profile("")
		if err != nil {
			t.Error(err)
			return
		}

		if _, err := s.GetKey("aws_api_key_duration"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("default-profile-env-var", func(t *testing.T) {
		os.Setenv(DefaultProfileEnvVar, "other")
		defer os.Unsetenv(DefaultProfileEnvVar)

		s, err := f.Profile("")
		if err != nil {
			t.Error(err)
			return
		}

		if _, err := s.GetKey("aws_api_key_duration"); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("missing-default", func(t *testing.T) {
		data := []byte("[s1]\nkey = val\n")
		f, err := Load(data)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Profile(""); err == nil {
			t.Error("loaded a non-existent default profile")
			return
		}
	})
}
