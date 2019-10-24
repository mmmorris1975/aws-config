package config

import (
	"os"
	"testing"
)

var p = NewEnvCredentialProvider()

func TestEnvCredentialProvider_Credentials(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		_, err := p.Credentials()
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("access key only", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", "myaccesskey")
		defer os.Unsetenv("AWS_ACCESS_KEY_ID")

		_, err := p.Credentials()
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("secret key only", func(t *testing.T) {
		os.Setenv("AWS_SECRET_ACCESS_KEY", "mysecret")
		defer os.Unsetenv("AWS_SECRET_ACCESS_KEY")

		_, err := p.Credentials()
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("access and secret key", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", "myaccesskey")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "mysecretkey")
		defer func() {
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
		}()

		c, err := p.Credentials()
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "myaccesskey" || c.SecretAccessKey != "mysecretkey" {
			t.Error("credential mismatch")
			return
		}
	})

	t.Run("everything", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY_ID", "myaccesskey")
		os.Setenv("AWS_SECRET_ACCESS_KEY", "mysecretkey")
		os.Setenv("AWS_SESSION_TOKEN", "mysessiontoken")
		defer func() {
			os.Unsetenv("AWS_ACCESS_KEY_ID")
			os.Unsetenv("AWS_SECRET_ACCESS_KEY")
			os.Unsetenv("AWS_SESSION_TOKEN")
		}()

		c, err := p.Credentials()
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "myaccesskey" || c.SecretAccessKey != "mysecretkey" || c.SessionToken != "mysessiontoken" {
			t.Error("credential mismatch")
			return
		}
	})

	t.Run("alternate vars", func(t *testing.T) {
		os.Setenv("AWS_ACCESS_KEY", "myaccesskey")
		os.Setenv("AWS_SECRET_KEY", "mysecretkey")
		defer func() {
			os.Unsetenv("AWS_ACCESS_KEY")
			os.Unsetenv("AWS_SECRET_KEY")
		}()

		c, err := p.Credentials()
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "myaccesskey" || c.SecretAccessKey != "mysecretkey" {
			t.Error("credential mismatch")
			return
		}
	})
}
