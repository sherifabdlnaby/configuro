package configuro

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

//Load load config into supported struct.
func (c *Config) Load(configStruct interface{}) error {
	var err error

	// Bind Env Vars
	if c.envLoad {
		c.bindAllEnvsWithPrefix()
	}

	if c.configFileLoad {
		err := c.viper.ReadInConfig()
		if err != nil {
			pathErr, ok := err.(*os.PathError)
			if !ok {
				return fmt.Errorf("error reading config data: %v", err)
			}

			if c.configFileErrIfNotFound && pathErr.Op == "open" {
				return fmt.Errorf("error config file not found. err: %v", err)
			}
		}
	}

	// Unmarshalling
	err = c.viper.Unmarshal(configStruct, c.decodeHook, setTagName(c.tag))

	if err != nil {
		return fmt.Errorf("error unmarshalling config: %v", err)
	}

	return nil
}

func (c *Config) bindAllEnvsWithPrefix() {
	envKVRegex := regexp.MustCompile("^" + c.envPrefix + "_" + "(.*)=.*$")
	Envvars := os.Environ()
	for _, env := range Envvars {
		match := envKVRegex.FindSubmatch([]byte(env))
		if match != nil {
			matchUnescaper := strings.NewReplacer("__", "_", "_", ".")
			matchUnescaped := matchUnescaper.Replace(string(match[1]))
			err := c.viper.BindEnv(matchUnescaped)

			if err != nil {
				//Should never happen tho.
				panic(err)
			}
		}
	}
}

func setTagName(hook string) viper.DecoderConfigOption {
	return func(c *mapstructure.DecoderConfig) {
		c.TagName = hook
	}
}

//StringJSONArrayOrSlicesToConfig will convert Json Encoded Strings to Maps or Slices, Used Primarily to support Slices and Maps in Environment variables
func stringJSONArrayToSlice() func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		var ret interface{}
		if t == reflect.Slice {
			jsonArray := make([]interface{}, 0)
			err := json.Unmarshal([]byte(raw), &jsonArray)
			if err != nil {
				// Try comma separated format too
				val, err := mapstructure.StringToSliceHookFunc(",").(func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error))(f, t, data)
				if err != nil {
					return val, err
				}
				ret = val
			} else {
				ret = jsonArray
			}
		}

		return ret, nil
	}
}

func stringJSONObjToMap() func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.Map {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		var ret interface{}
		jsonMap := make(map[string]interface{})
		err := json.Unmarshal([]byte(raw), &jsonMap)
		if err != nil {
			return raw, fmt.Errorf("couldn't map string-ifed Json to Map: %s", err.Error())
		}
		ret = jsonMap

		return ret, nil
	}
}

func stringJSONObjToStruct() func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || t != reflect.Struct {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		var ret interface{}
		err := json.Unmarshal([]byte(raw), &data)
		if err != nil {
			return raw, fmt.Errorf("couldn't map string-ifed Json to Object: %s", err.Error())
		}
		ret = data
		return ret, nil
	}
}

func expandEnvVariablesWithDefaults() func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	var configWithEnvExpand = regexp.MustCompile(`(\${([\w@.]+)(\|([\w@.]+)?)?})`)
	var exactMatchEnvExpand = regexp.MustCompile(`^` + configWithEnvExpand.String() + `$`)
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String {
			return data, nil
		}

		raw := data.(string)
		ret := configWithEnvExpand.ReplaceAllStringFunc(raw, func(s string) string {
			matches := exactMatchEnvExpand.FindAllStringSubmatch(s, -1)
			if matches == nil {
				return s
			}
			envKey := matches[0][2]
			isEnvDefaultSet := matches[0][3] != ""
			envDefault := matches[0][4]
			envValue, found := os.LookupEnv(envKey)
			if !found {
				if isEnvDefaultSet {
					return envDefault
				}
				return s
			}
			return envValue
		})
		return ret, nil
	}
}
