package config

import (
	"fmt"
	"github.com/go-ini/ini"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	// ProfileEnvVar is the profile environment variable name
	ProfileEnvVar = "AWS_PROFILE"
	// DefaultProfileEnvVar is the default profile environment variable name
	DefaultProfileEnvVar = "AWS_DEFAULT_PROFILE"
)

// DefaultProfileName is the name of the default section in the config file
var DefaultProfileName = strings.ToLower(ini.DefaultSection)

type awsConfigFile struct {
	*ini.File
	Path   string
	isTemp bool
}

func load(source interface{}, def func(f *awsConfigFile)) (*awsConfigFile, error) {
	f := new(awsConfigFile)

	switch t := source.(type) {
	case string:
		// path to local file, or url (file and http(s) supported)
		if u, err := url.Parse(t); err == nil {
			if err := f.urlHandler(u); err != nil {
				return nil, err
			}
			source = f.Path
		}
	case *url.URL:
		if err := f.urlHandler(t); err != nil {
			return nil, err
		}
		source = f.Path
	case []byte:
		// raw bytes, explicitly supported in go-ini
		f.isTemp = false
	case *os.File:
		// file object, explicitly supported in go-ini (just set path attribute in our struct)
		f.Path = t.Name()
		f.isTemp = false
	case io.Reader:
		// other kind of reader (ensure it's a ReadCloser here so it's supported by go-ini)
		if _, ok := t.(io.ReadCloser); !ok {
			source = ioutil.NopCloser(t)
		}
		f.isTemp = false
	default:
		source = []byte("[default]")

		// callback to perform default action
		if def != nil {
			def(f)
		}

		if len(f.Path) > 0 {
			if _, err := os.Stat(f.Path); err == nil {
				source = f.Path
			}
		}
	}

	s, err := ini.Load(source)
	if err != nil {
		return nil, err
	}
	f.File = s

	return f, nil
}

func (f *awsConfigFile) ProfileStrings() []string {
	s := make([]string, 0)
	fmt.Printf("%+v\n", f.SectionStrings())
	for _, v := range f.SectionStrings() {
		// Skip the go-ini DEFAULT section
		if v != ini.DefaultSection {
			s = append(s, strings.TrimPrefix(v, "profile "))
		}
	}
	return s
}

// Default profile resolution method.  Returns an error if the section is not found in the config file.
// Since there is no explicit format of the profile data (the cli/sdk defines some values, but allows custom attributes),
// the returned value is the ini file section data, where the caller will be able to process the profile attributes locally.
func (f *awsConfigFile) Profile(name string) (*ini.Section, error) {
	return f.profile(name, nil)
}

// Attempt to fetch the given profile name.  If an empty string is passed, check env vars, or provide the default name.
// An optional function can be provided as a handler to return an alternate profile name, in case the original name is
// not found.  This should satisfy the oddity that is the AWS config file where non-default profiles should be prefixed
// with "profile" in the name.
func (f *awsConfigFile) profile(name string, nfh func(n string) string) (*ini.Section, error) {
	name = ResolveProfile(&name)

	s, err := f.GetSection(name)
	if err != nil {
		if nfh != nil {
			return f.GetSection(nfh(name))
		}
		return nil, err
	}
	return s, nil
}

func (f *awsConfigFile) Close() error {
	if f.isTemp {
		return os.Remove(f.Path)
	}
	return nil
}

func (f *awsConfigFile) urlHandler(u *url.URL) error {
	switch u.Scheme {
	case "http", "https":
		tf, err := fetchHttpSource(u)
		if err != nil {
			return err
		}
		f.Path = tf.Name()
		f.isTemp = true
	case "file":
		f.Path = u.Opaque
		f.isTemp = false
	case "":
		f.Path = u.Path
		f.isTemp = false
	default:
		// error: not supported
		return fmt.Errorf("url scheme '%s' not supported", u.Scheme)
	}
	return nil
}

func fetchHttpSource(u *url.URL) (*os.File, error) {
	r, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP Response Code %d", r.StatusCode)
	}

	return ioutil.TempFile("", "AwsConfigLoader-")
}

// ResolveProfile is a helper method to check the env vars for a profile name if the provided argument is nil or empty
func ResolveProfile(p *string) string {
	if p == nil || len(*p) < 1 {
		var n string
		var ok bool
		if n, ok = os.LookupEnv(ProfileEnvVar); !ok {
			if n, ok = os.LookupEnv(DefaultProfileEnvVar); !ok {
				n = DefaultProfileName
			}
		}
		return n
	}

	return *p
}
