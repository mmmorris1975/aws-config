package config

import (
	"bytes"
	"context"
	"net/http"
	"net/url"
	"os"
	"testing"
)

var ConfFileName = ".aws_config"

func TestLoad(t *testing.T) {
	t.Run("LoadString", func(t *testing.T) {
		t.Run("GoodPath", func(t *testing.T) {
			c, err := load(ConfFileName, nil)
			if err != nil {
				t.Error(err)
			}

			if c.Path != ConfFileName {
				t.Error("file name mismatch")
			}

			if len(c.Sections()) < 1 {
				t.Error("missing section data")
			}
		})

		t.Run("BadPath", func(t *testing.T) {
			if _, err := load("not_my_file", nil); err == nil {
				t.Error("did not receive expected error with bad file path")
				return
			}
		})

		t.Run("FileUrl", func(t *testing.T) {
			c, err := load(ConfFileName, nil)
			if err != nil {
				t.Error(err)
				return
			}

			if c.Path != ConfFileName {
				t.Error("file name mismatch")
			}

			if len(c.Sections()) < 1 {
				t.Error("missing section data")
			}
		})
	})

	t.Run("LoadURL", func(t *testing.T) {
		svr := http.Server{Addr: ":8888", Handler: http.FileServer(http.Dir("."))}
		go func() {
			svr.ListenAndServe()
		}()
		defer svr.Shutdown(context.Background())

		t.Run("Path", func(t *testing.T) {
			u := url.URL{Path: ConfFileName}
			c, err := load(&u, nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer c.Close()

			if c.Path != ConfFileName {
				t.Error("file name mismatch")
			}

			if len(c.Sections()) < 1 {
				t.Error("missing section data")
			}
		})

		t.Run("FileUrl", func(t *testing.T) {
			u := url.URL{Scheme: "file", Opaque: ConfFileName}
			c, err := load(&u, nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer c.Close()

			if c.Path != ConfFileName {
				t.Error("file name mismatch")
			}

			if len(c.Sections()) < 1 {
				t.Error("missing section data")
			}
		})

		t.Run("HttpUrlGood", func(t *testing.T) {
			u, _ := url.Parse("http://localhost:8888/.aws_config")
			c, err := load(u, nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer c.Close()

			if len(c.Sections()) < 1 {
				t.Error("missing section data")
			}
		})

		t.Run("HttpUrlBad", func(t *testing.T) {
			u, _ := url.Parse("http://localhost:8888/not_my_file")
			if _, err := load(u, nil); err == nil {
				t.Error("did not receive expected error with bad HTTP url")
				return
			}
		})

		t.Run("FtpUrl", func(t *testing.T) {
			u := url.URL{Scheme: "ftp", Host: "localhost", Path: ConfFileName}
			if _, err := load(&u, nil); err == nil {
				t.Error("did not receive expected error with an FTP url")
				return
			}
		})
	})

	t.Run("LoadFile", func(t *testing.T) {
		f, err := os.Open(ConfFileName)
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()

		buf := new(bytes.Buffer)
		if _, err := buf.ReadFrom(f); err != nil {
			t.Error(err)
			return
		}

		t.Run("Bytes", func(t *testing.T) {
			c, err := load(buf.Bytes(), nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer c.Close()

			if len(c.Sections()) < 1 {
				t.Error("missing section data")
			}
		})

		t.Run("Reader", func(t *testing.T) {
			c, err := load(buf, nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer c.Close()

			if len(c.Sections()) < 1 {
				t.Error("missing section data")
			}
		})

		t.Run("FileObj", func(t *testing.T) {
			f.Seek(0, 0)
			c, err := load(f, nil)
			if err != nil {
				t.Error(err)
				return
			}
			defer c.Close()

			if c.Path != ConfFileName {
				t.Error("file name mismatch")
			}

			if len(c.Sections()) < 1 {
				t.Error("missing section data")
			}
		})
	})

	t.Run("DefaultNoFallback", func(t *testing.T) {
		f, err := load(nil, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer f.Close()
	})
}

func TestProfileStrings(t *testing.T) {
	c, err := load(ConfFileName, nil)
	if err != nil {
		t.Error(err)
		return
	}

	if len(c.ProfileStrings()) < 1 {
		t.Error("missing profiles")
		return
	}
	t.Log(c.ProfileStrings())

	if c.ProfileStrings()[0] != DefaultProfileName {
		t.Error("default profile not found")
		return
	}
}

func TestProfile(t *testing.T) {
	c, err := load(ConfFileName, nil)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("EmptyDefault", func(t *testing.T) {
		s, err := c.Profile("")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != DefaultProfileName {
			t.Error("profile name mismatch")
		}
	})

	t.Run("ExplicitDefault", func(t *testing.T) {
		s, err := c.Profile("default")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != DefaultProfileName {
			t.Error("profile name mismatch")
		}
	})

	t.Run("ProfileEnv", func(t *testing.T) {
		os.Setenv(ProfileEnvVar, DefaultProfileName)
		defer os.Unsetenv(ProfileEnvVar)

		s, err := c.Profile("")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != DefaultProfileName {
			t.Error("profile name mismatch")
		}
	})

	t.Run("DefaultProfileEnv", func(t *testing.T) {
		os.Setenv(DefaultProfileEnvVar, DefaultProfileName)
		defer os.Unsetenv(DefaultProfileEnvVar)

		s, err := c.Profile("")
		if err != nil {
			t.Error(err)
			return
		}

		if s.Name() != DefaultProfileName {
			t.Error("profile name mismatch")
		}
	})
}
