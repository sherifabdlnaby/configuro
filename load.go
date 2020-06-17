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

	if c.configFileLoad {
		err := c.viper.ReadInConfig()
		if err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("error reading config data: %v", err)
			}
		}
	}

	// BindEnvvars
	c.bindEnvs(configStruct)

	// Unmarshalling
	err = c.viper.Unmarshal(configStruct, c.decodeHook, setTagName(c.tag))

	if err != nil {
		return fmt.Errorf("error unmarshalling config: %v", err)
	}

	return nil
}

func (c *Config) bindEnvs(iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	if ift.Kind() == reflect.Ptr {
		ift = ift.Elem()
		ifv = reflect.Indirect(ifv)
	}
	for i := 0; i < ift.NumField(); i++ {
		fieldv := ifv.Field(i)
		if !fieldv.CanInterface() {
			return
		}
		t := ift.Field(i)
		name := strings.ToLower(t.Name)
		tag, ok := t.Tag.Lookup(c.tag)
		if ok {
			name = tag
		}
		path := append(parts, name)
		switch fieldv.Kind() {
		case reflect.Struct:
			c.bindEnvs(fieldv.Interface(), path...)
		default:
			_ = c.viper.BindEnv(strings.Join(path, "."))
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

func stringJSONObjToMapOrStruct() func(f reflect.Kind, t reflect.Kind, data interface{}) (interface{}, error) {
	return func(
		f reflect.Kind,
		t reflect.Kind,
		data interface{}) (interface{}, error) {
		if f != reflect.String || (t != reflect.Map && t != reflect.Struct) {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		var ret interface{}
		if t == reflect.Map {
			jsonMap := make(map[string]interface{})
			err := json.Unmarshal([]byte(raw), &jsonMap)
			if err != nil {
				return raw, fmt.Errorf("couldn't map string-ifed Json to Map: %s", err.Error())
			}
			ret = jsonMap
		} else if t == reflect.Struct {
			err := json.Unmarshal([]byte(raw), &data)
			if err != nil {
				return raw, fmt.Errorf("couldn't map string-ifed Json to Object: %s", err.Error())
			}
			ret = data
		}
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
