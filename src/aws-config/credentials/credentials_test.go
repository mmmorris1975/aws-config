package credentials

import (
	. "aws-config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/iam"
	"os"
	"testing"
)

const CredFile = "test-creds"

func TestLoad(t *testing.T) {
	t.Run("valid-string", func(t *testing.T) {
		if _, err := Load(CredFile); err != nil {
			t.Error("failed to load credentials file")
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
		os.Setenv(CredFileEnvVar, CredFile)
		defer os.Unsetenv(CredFileEnvVar)

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
				t.Error("panic when trying to load default cred file")
				return
			}
		}()

		Load(nil)
	})
}

func TestAwsCredentialsFile_Credentials(t *testing.T) {
	t.Run("explicit-default", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		c, err := f.Credentials("default")
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "AKIAM0CK" {
			t.Error("mismatched access key")
			return
		}
	})

	t.Run("empty", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		c, err := f.Credentials("")
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "AKIAM0CK" {
			t.Error("mismatched access key")
			return
		}
	})

	t.Run("non-default", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		c, err := f.Credentials("other")
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "AKIA0THER" {
			t.Error("mismatched access key")
			return
		}
	})

	t.Run("missing-section", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Credentials("missing"); err == nil {
			t.Error("successfully loaded a missing section")
			return
		}
	})

	t.Run("empty-section", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Credentials("empty"); err == nil {
			t.Error("did not see error when loading empty section")
			return
		}
	})

	t.Run("missing-accesskey", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Credentials("no-access"); err == nil {
			t.Error("did not see error when loading incomplete section")
			return
		}
	})

	t.Run("missing-secretkey", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Credentials("no-secret"); err == nil {
			t.Error("did not see error when loading incomplete section")
			return
		}
	})

	t.Run("bad-properties", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Credentials("bad-props"); err == nil {
			t.Error("did not see error when loading bad section")
			return
		}
	})

	t.Run("env-var", func(t *testing.T) {
		os.Setenv(ProfileEnvVar, "other")
		defer os.Unsetenv(ProfileEnvVar)

		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Credentials(""); err != nil {
			t.Error("failed to load credentials via env var")
			return
		}
	})

	t.Run("default-env-var", func(t *testing.T) {
		os.Setenv(DefaultProfileEnvVar, "other")
		defer os.Unsetenv(DefaultProfileEnvVar)

		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if _, err := f.Credentials(""); err != nil {
			t.Error("failed to load credentials via default env var")
			return
		}
	})
}

func TestAwsCredentialsFile_UpdateCredentials(t *testing.T) {
	t.Run("nil-creds", func(t *testing.T) {
		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.UpdateCredentials("x", nil); err == nil {
			t.Error("successfully set nil credentials")
			return
		}
	})

	t.Run("invalid-cred-type", func(t *testing.T) {
		c := make(map[string]string)
		c["aws_access_key_id"] = "MapKey"
		c["aws_secret_access_key"] = "MapSecret"

		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.UpdateCredentials("x", c); err == nil {
			t.Error("successfully set invalid credentials")
			return
		}
	})

	t.Run("access-key-type", func(t *testing.T) {
		c := iam.AccessKey{
			AccessKeyId:     aws.String("AKIAK3Y"),
			SecretAccessKey: aws.String("MOCK"),
			Status:          aws.String(iam.StatusTypeActive),
		}

		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("value", func(t *testing.T) {
			if err := f.UpdateCredentials("new-ak", c); err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("pointer", func(t *testing.T) {
			if err := f.UpdateCredentials("empty", &c); err != nil {
				t.Error(err)
				return
			}
		})
	})

	t.Run("credentials-type", func(t *testing.T) {
		c := credentials.Value{AccessKeyID: "AKIACR3D", SecretAccessKey: "M0CK"}

		f, err := Load(CredFile)
		if err != nil {
			t.Fatal(err)
		}

		t.Run("value", func(t *testing.T) {
			if err := f.UpdateCredentials("new-cred", c); err != nil {
				t.Error(err)
				return
			}
		})

		t.Run("pointer", func(t *testing.T) {
			if err := f.UpdateCredentials("empty", &c); err != nil {
				t.Error(err)
				return
			}
		})
	})
}
