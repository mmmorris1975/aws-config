package config

import (
	"os"
	"testing"
)

var cfg = NewEnvConfigProvider()

func TestEnvConfigProvider_Config(t *testing.T) {
	t.Run("profile and region", func(t *testing.T) {
		os.Setenv("AWS_PROFILE", "pfile")
		os.Setenv("AWS_REGION", "us-east-2")
		defer func() {
			os.Unsetenv("AWS_PROFILE")
			os.Unsetenv("AWS_REGION")
		}()

		c, err := cfg.Config()
		if err != nil {
			t.Error(err)
			return
		}

		if c.Profile != "pfile" || c.Region != "us-east-2" {
			t.Error("data mismatch")
		}
	})

	t.Run("mfa", func(t *testing.T) {
		os.Setenv("MFA_SERIAL", "arn:aws::iam:mfa/my-mfa")
		os.Setenv("EXTERNAL_ID", "ext-id-12345")
		os.Setenv("DURATION_SECONDS", "14400")
		defer func() {
			os.Unsetenv("MFA_SERIAL")
			os.Unsetenv("EXTERNAL_ID")
			os.Unsetenv("DURATION_SECONDS")
		}()

		c, err := cfg.Config()
		if err != nil {
			t.Error(err)
			return
		}

		if len(c.MfaSerial) < 1 || c.ExternalId != "ext-id-12345" || c.DurationSeconds != 14400 {
			t.Error("data mismatch")
		}
	})

	t.Run("profile override", func(t *testing.T) {
		os.Setenv("AWS_PROFILE", "profile1")
		defer os.Unsetenv("AWS_PROFILE")

		c, err := cfg.Config("profile2")
		if err != nil {
			t.Error(err)
			return
		}

		if c.Profile != "profile2" {
			t.Error("incorrect profile")
		}
	})

	t.Run("credential duration", func(t *testing.T) {
		os.Setenv("CREDENTIALS_DURATION", "10h10m10s")
		defer os.Unsetenv("CREDENTIALS_DURATION")

		c, err := cfg.Config()
		if err != nil {
			t.Error(err)
			return
		}

		if c.DurationSeconds != 36610 {
			t.Errorf("bad duration: %d", c.DurationSeconds)
		}
	})

	t.Run("alternate env vars", func(t *testing.T) {
		os.Setenv("AWS_DEFAULT_REGION", "us-west-1")
		defer os.Unsetenv("AWS_DEFAULT_REGION")

		c, err := cfg.Config()
		if err != nil {
			t.Error(err)
			return
		}

		if c.Region != "us-west-1" {
			t.Error("bad region")
		}
	})
}
