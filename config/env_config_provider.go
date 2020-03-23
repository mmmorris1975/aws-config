package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// EnvConfigProvider enables the lookup of AWS configuration from environment variables
type EnvConfigProvider uint8

// NewEnvConfigProvider creates an EnvConfigProvider with the default configuration
func NewEnvConfigProvider() *EnvConfigProvider {
	return new(EnvConfigProvider)
}

// Config will return the configuration attributes found in the environment variables.  The profile argument to this
// call is ignored, and only used to set the Profile attribute of the returned AwsConfig object.
func (p *EnvConfigProvider) Config(profile ...string) (*AwsConfig, error) {
	c := AwsConfig{}
	c.rawAttributes = make(map[string]string)

	v := reflect.ValueOf(&c)
	t := reflect.TypeOf(c)
	for i := 0; i < t.NumField(); i++ {
		tField := t.Field(i)
		vField := v.Elem().Field(i)

		e := lookupEnvTag(tField.Tag.Get("env"))

		// set value in rawAttributes as well, if the env var is set, and there's a corresponding ini tag
		// makes sure that Merge() in the config resolver works as expected
		if f := tField.Tag.Get("ini"); len(f) > 0  && len(e) > 0 {
			c.rawAttributes[f] = e
		}

		switch tField.Type.Kind() {
		case reflect.String:
			vField.SetString(e)
		case reflect.Bool:
			b, err := strconv.ParseBool(e)
			if err != nil {
				b = false
			}
			vField.SetBool(b)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(e, 0, 64)
			if err != nil {
				i = 0
			}
			vField.SetInt(i)
			//case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			//	// byte is an alias for uint8, so supporting uint8 breaks support for byte
			//	i, err := strconv.ParseUint(e, 0, 64)
			//	if err != nil {
			//		i = 0
			//	}
			//	vField.SetUint(i)
			//case reflect.Float32, reflect.Float64:
			//	f, err := strconv.ParseFloat(e, 64)
			//	if err != nil {
			//		f = 0
			//	}
			//	vField.SetFloat(f)
		}
	}

	if profile != nil && len(profile) > 0 {
		c.Profile = profile[0]
	}

	return &c, nil
}

// ListProfiles is not supported for EnvConfigProviders, returns an empty array
func (p *EnvConfigProvider) ListProfiles(roles bool) []string {
	return []string{}
}

func lookupEnvTag(tag string) string {
	if len(tag) > 0 {
		for _, s := range strings.Split(tag, ",") {
			if v, ok := os.LookupEnv(s); ok {
				if s == "CREDENTIALS_DURATION" {
					d, err := time.ParseDuration(v)
					if err != nil {
						continue
					}
					v = fmt.Sprintf("%d", int64(d.Seconds()))
				}

				return v
			}
		}
	}
	return ""
}