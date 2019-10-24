package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type EnvConfigProvider uint8

func NewEnvConfigProvider() *EnvConfigProvider {
	return new(EnvConfigProvider)
}

func (p *EnvConfigProvider) Config(profile ...string) (*AwsConfig, error) {
	c := AwsConfig{}

	v := reflect.ValueOf(&c)
	t := reflect.TypeOf(c)
	for i := 0; i < t.NumField(); i++ {
		tField := t.Field(i)
		vField := v.Elem().Field(i)

		e := lookupEnvTag(tField.Tag.Get("env"))

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
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// byte is an alias for uint8, so supporting uint8 breaks support for byte
			i, err := strconv.ParseUint(e, 0, 64)
			if err != nil {
				i = 0
			}
			vField.SetUint(i)
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(e, 64)
			if err != nil {
				f = 0
			}
			vField.SetFloat(f)
		}
	}

	if profile != nil && len(profile) > 0 {
		c.Profile = profile[0]
	}

	return &c, nil
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
