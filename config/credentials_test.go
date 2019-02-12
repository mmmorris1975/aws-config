package config

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/iam"
	"os"
	"testing"
)

var CredFileName = ".aws_credentials"

func TestNewAwsCredentialsFile(t *testing.T) {
	t.Run("ExplicitSource", func(t *testing.T) {
		f, err := NewAwsCredentialsFile(CredFileName)
		if err != nil {
			t.Error(err)
			return
		}

		if f.Path != CredFileName {
			t.Error("file path mismatch")
			return
		}
	})

	t.Run("EnvVarSource", func(t *testing.T) {
		os.Setenv(CredentialsFileEnvVar, CredFileName)
		defer os.Unsetenv(CredentialsFileEnvVar)

		f, err := NewAwsCredentialsFile(nil)
		if err != nil {
			t.Error(err)
			return
		}

		if f.Path != CredFileName {
			t.Error("file path mismatch")
			return
		}
	})

	t.Run("BadSource", func(t *testing.T) {
		_, err := NewAwsCredentialsFile("not-a-file")
		if err == nil {
			t.Error("did not get expected error")
			return
		}
	})
}

func TestAwsCredentialsFile_Credentials(t *testing.T) {
	f, err := NewAwsCredentialsFile(CredFileName)
	if err != nil {
		t.Error("did not get expected error")
		return
	}
	defer f.Close()

	t.Run("MissingProfile", func(t *testing.T) {
		c, err := f.Credentials("")
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "AKIAM0CK" {
			t.Error("credentials mismatch")
			return
		}
	})

	t.Run("GoodProfile", func(t *testing.T) {
		c, err := f.Credentials("other")
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "AKIA0THER" {
			t.Error("credentials mismatch")
			return
		}
	})

	t.Run("BadProfile", func(t *testing.T) {
		_, err := f.Credentials("nope")
		if err == nil {
			t.Error("did not see expected error")
			return
		}
	})

	t.Run("EmptyProfile", func(t *testing.T) {
		_, err := f.Credentials("empty")
		if err == nil {
			t.Error("did not see expected error")
			return
		}
	})

	t.Run("MissingSecret", func(t *testing.T) {
		_, err := f.Credentials("no-secret")
		if err == nil {
			t.Error("did not see expected error")
			return
		}
	})

	t.Run("MissingAccess", func(t *testing.T) {
		_, err := f.Credentials("no-access")
		if err == nil {
			t.Error("did not see expected error")
			return
		}
	})

	t.Run("MissingProps", func(t *testing.T) {
		_, err := f.Credentials("bad-props")
		if err == nil {
			t.Error("did not see expected error")
			return
		}
	})
}

func TestAwsCredentialsFile_UpdateCredentials(t *testing.T) {
	f, err := NewAwsCredentialsFile(CredFileName)
	if err != nil {
		t.Error("did not get expected error")
		return
	}
	defer f.Close()

	t.Run("NilCreds", func(t *testing.T) {
		if err := f.UpdateCredentials("default", nil); err == nil {
			t.Error("did not see expected error")
			return
		}
	})

	t.Run("StringCreds", func(t *testing.T) {
		if err := f.UpdateCredentials("default", "cred/string"); err == nil {
			t.Error("did not see expected error")
			return
		}
	})

	t.Run("IamCredsVal", func(t *testing.T) {
		c := iam.AccessKey{AccessKeyId: aws.String("xxx"), SecretAccessKey: aws.String("yyy")}

		t.Run("Active", func(t *testing.T) {
			c.Status = aws.String(iam.StatusTypeActive)
			if err := f.UpdateCredentials("other", c); err != nil {
				t.Error(err)
				return
			}

			cr, err := f.Credentials("other")
			if err != nil {
				t.Error(err)
				return
			}

			if cr.AccessKeyID != *c.AccessKeyId {
				t.Error("access key mismatch")
				return
			}
		})

		t.Run("Inactive", func(t *testing.T) {
			c.Status = aws.String(iam.StatusTypeInactive)
			if err := f.UpdateCredentials("default", c); err != nil {
				t.Error(err)
				return
			}

			cr, err := f.Credentials("default")
			if err != nil {
				t.Error(err)
				return
			}

			if cr.AccessKeyID == *c.AccessKeyId {
				t.Error("access key unexpectedly updated")
				return
			}
		})
	})

	t.Run("IamCredsPtr", func(t *testing.T) {
		c := &iam.AccessKey{AccessKeyId: aws.String("abc"), SecretAccessKey: aws.String("123")}

		t.Run("Active", func(t *testing.T) {
			c.Status = aws.String(iam.StatusTypeActive)
			if err := f.UpdateCredentials("other", c); err != nil {
				t.Error(err)
				return
			}

			cr, err := f.Credentials("other")
			if err != nil {
				t.Error(err)
				return
			}

			if cr.AccessKeyID != *c.AccessKeyId {
				t.Error("access key mismatch")
				return
			}
		})

		t.Run("Inactive", func(t *testing.T) {
			c.Status = aws.String(iam.StatusTypeInactive)
			if err := f.UpdateCredentials("default", c); err != nil {
				t.Error(err)
				return
			}

			cr, err := f.Credentials("default")
			if err != nil {
				t.Error(err)
				return
			}

			if cr.AccessKeyID == *c.AccessKeyId {
				t.Error("access key unexpectedly updated")
				return
			}
		})
	})

	t.Run("CredentialsVal", func(t *testing.T) {
		c := credentials.Value{AccessKeyID: "aaa", SecretAccessKey: "bbb"}
		if err := f.UpdateCredentials("other", c); err != nil {
			t.Error(err)
			return
		}

		cr, err := f.Credentials("other")
		if err != nil {
			t.Error(err)
			return
		}

		if cr.AccessKeyID != c.AccessKeyID {
			t.Error("access key mismatch")
			return
		}
	})

	t.Run("CredentialsPtr", func(t *testing.T) {
		c := &credentials.Value{AccessKeyID: "ccc", SecretAccessKey: "ddd"}
		if err := f.UpdateCredentials("other", c); err != nil {
			t.Error(err)
			return
		}

		cr, err := f.Credentials("other")
		if err != nil {
			t.Error(err)
			return
		}

		if cr.AccessKeyID != c.AccessKeyID {
			t.Error("access key mismatch")
			return
		}
	})
}
