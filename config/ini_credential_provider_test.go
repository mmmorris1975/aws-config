package config

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/iam"
	"io/ioutil"
	"net/url"
	"os"
	"testing"
)

var credFileName = ".aws_credentials"

func TestNewIniCredentialProvider(t *testing.T) {
	t.Run("nil source", func(t *testing.T) {
		os.Setenv(CredentialsFileEnvVar, os.DevNull)
		defer os.Unsetenv(CredentialsFileEnvVar)

		p, err := NewIniCredentialProvider(nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer p.Close()

		if len(p.ProfileStrings()) > 1 {
			t.Error("unexpected profiles found")
		}
	})

	t.Run("string source", func(t *testing.T) {
		p, err := NewIniCredentialProvider(credFileName)
		if err != nil {
			t.Error(err)
			return
		}
		p.Close()
	})

	t.Run("file url string source", func(t *testing.T) {
		p, err := NewIniCredentialProvider("file:" + credFileName)
		if err != nil {
			t.Error(err)
			return
		}
		p.Close()
	})

	t.Run("file url.URL source", func(t *testing.T) {
		u, err := url.Parse("file:" + credFileName)
		if err != nil {
			t.Error(err)
			return
		}

		p, err := NewIniCredentialProvider(u)
		if err != nil {
			t.Error(err)
			return
		}
		p.Close()
	})

	t.Run("byte source", func(t *testing.T) {
		b, err := ioutil.ReadFile(credFileName)
		if err != nil {
			t.Error(err)
			return
		}

		p, err := NewIniCredentialProvider(b)
		if err != nil {
			t.Error(err)
			return
		}
		p.Close()
	})

	t.Run("os.File source", func(t *testing.T) {
		f, err := os.Open(credFileName)
		if err != nil {
			t.Error(err)
			return
		}

		p, err := NewIniCredentialProvider(f)
		if err != nil {
			t.Error(err)
			return
		}
		p.Close()
	})

	t.Run("io.Reader source", func(t *testing.T) {
		b, err := ioutil.ReadFile(credFileName)
		if err != nil {
			t.Error(err)
			return
		}

		p, err := NewIniCredentialProvider(bytes.NewReader(b))
		if err != nil {
			t.Error(err)
			return
		}
		p.Close()
	})

	t.Run("env var source", func(t *testing.T) {
		os.Setenv(CredentialsFileEnvVar, credFileName)
		defer os.Unsetenv(CredentialsFileEnvVar)

		p, err := NewIniCredentialProvider(nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer p.Close()

		if p.Path != credFileName {
			t.Error("file name mismatch")
			return
		}
	})

	t.Run("bad source", func(t *testing.T) {
		_, err := NewIniCredentialProvider("not-a-file")
		if err == nil {
			t.Error("did not receive expected error")
			return
		}
	})
}

func TestIniCredentialProvider_Credentials(t *testing.T) {
	f, err := NewIniCredentialProvider(credFileName)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	// This should return the default profile credentials
	t.Run("empty profile", func(t *testing.T) {
		c, err := f.Credentials()
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "AKIAM0CK" {
			t.Error("bad access key")
		}

		if c.SecretAccessKey != "M0cKSecr3T" {
			t.Error("bad secret key")
		}
	})

	t.Run("simple profile", func(t *testing.T) {
		c, err := f.Credentials("other")
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "AKIA0THER" {
			t.Error("bad access key")
		}

		if c.SecretAccessKey != "0th3rSecr3T" {
			t.Error("bad secret key")
		}
	})

	t.Run("session token profile", func(t *testing.T) {
		c, err := f.Credentials("token")
		if err != nil {
			t.Error(err)
			return
		}

		if c.AccessKeyID != "accesskey" {
			t.Error("bad access key")
		}

		if c.SecretAccessKey != "secretkey" {
			t.Error("bad secret key")
		}

		if c.SessionToken != "sessioncreds" {
			t.Error("bad session token")
		}
	})

	t.Run("missing profile", func(t *testing.T) {
		_, err := f.Credentials("nope")
		if err == nil {
			t.Error("did not receive expected error")
		}
	})

	t.Run("empty profile", func(t *testing.T) {
		_, err := f.Credentials("empty")
		if err == nil {
			t.Error("did not receive expected error")
		}
	})

	t.Run("missing secret key", func(t *testing.T) {
		_, err := f.Credentials("no-secret")
		if err == nil {
			t.Error("did not receive expected error")
		}
	})

	t.Run("missing access key", func(t *testing.T) {
		_, err := f.Credentials("no-access")
		if err == nil {
			t.Error("did not receive expected error")
		}
	})

	t.Run("bad properties", func(t *testing.T) {
		_, err := f.Credentials("bad-props")
		if err == nil {
			t.Error("did not receive expected error")
		}
	})
}

func TestIniCredentialProvider_UpdateCredentials(t *testing.T) {
	f, err := NewIniCredentialProvider(credFileName)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()

	// This is a no-op
	t.Run("nil creds", func(t *testing.T) {
		if err := f.UpdateCredentials(DefaultProfileName, nil); err != nil {
			t.Error(err)
			return
		}
	})

	t.Run("string creds", func(t *testing.T) {
		if err := f.UpdateCredentials(DefaultProfileName, "string/creds"); err == nil {
			t.Error("did not receive expected error")
			return
		}
	})

	t.Run("iam creds", func(t *testing.T) {
		c := iam.AccessKey{
			AccessKeyId:     aws.String("accesskey"),
			SecretAccessKey: aws.String("secretkey"),
		}

		t.Run("by value", func(t *testing.T) {
			t.Run("active", func(t *testing.T) {
				c.Status = aws.String(iam.StatusTypeActive)
				if err := f.UpdateCredentials("other", c); err != nil {
					t.Error(err)
					return
				}

				v, err := f.Credentials("other")
				if err != nil {
					t.Error(err)
					return
				}

				if *c.AccessKeyId != v.AccessKeyID {
					t.Error("access key mismatch")
				}
			})

			t.Run("inactive", func(t *testing.T) {
				c.Status = aws.String(iam.StatusTypeInactive)
				if err := f.UpdateCredentials(DefaultProfileName, c); err != nil {
					t.Error(err)
					return
				}

				v, err := f.Credentials(DefaultProfileName)
				if err != nil {
					t.Error(err)
					return
				}

				if *c.AccessKeyId == v.AccessKeyID {
					t.Error("access key unexpectedly updated")
				}
			})
		})

		t.Run("by reference", func(t *testing.T) {
			t.Run("active", func(t *testing.T) {
				c.Status = aws.String(iam.StatusTypeActive)
				if err := f.UpdateCredentials("other", &c); err != nil {
					t.Error(err)
					return
				}

				v, err := f.Credentials("other")
				if err != nil {
					t.Error(err)
					return
				}

				if *c.AccessKeyId != v.AccessKeyID {
					t.Error("access key mismatch")
				}
			})

			t.Run("inactive", func(t *testing.T) {
				c.Status = aws.String(iam.StatusTypeInactive)
				if err := f.UpdateCredentials(DefaultProfileName, &c); err != nil {
					t.Error(err)
					return
				}

				v, err := f.Credentials(DefaultProfileName)
				if err != nil {
					t.Error(err)
					return
				}

				if *c.AccessKeyId == v.AccessKeyID {
					t.Error("access key unexpectedly updated")
				}
			})
		})
	})

	t.Run("credentials", func(t *testing.T) {
		c := credentials.Value{
			AccessKeyID:     "accesskey",
			SecretAccessKey: "secretkey",
		}

		t.Run("by value", func(t *testing.T) {
			if err := f.UpdateCredentials("other", c); err != nil {
				t.Error(err)
				return
			}

			v, err := f.Credentials("other")
			if err != nil {
				t.Error(err)
				return
			}

			if c.AccessKeyID != v.AccessKeyID {
				t.Error("access key mismatch")
			}
		})

		t.Run("by reference", func(t *testing.T) {
			if err := f.UpdateCredentials("other", &c); err != nil {
				t.Error(err)
				return
			}

			v, err := f.Credentials("other")
			if err != nil {
				t.Error(err)
				return
			}

			if c.AccessKeyID != v.AccessKeyID {
				t.Error("access key mismatch")
			}
		})
	})
}
