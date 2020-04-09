package configuro

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	ens "github.com/go-playground/validator/translations/en"
	"github.com/hashicorp/go-multierror"
	"github.com/joho/godotenv"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

type Error multierror.Error
type Config struct {
	envLoad                bool
	envPrefix              string
	envDotFileLoad         bool
	envDotFilePath         string
	configFileLoad         bool
	configFileName         string
	configFileDir          string
	configDirEnv           bool
	configDirEnvName       string
	configEnvExpand        bool
	validateStopOnFirstErr bool
	validateRecursive      bool
	validateUsingTags      bool
	tag                    string
	viper                  *viper.Viper
	validator              *validator.Validate
	validatorTrans         ut.Translator
	decodeHook             viper.DecoderConfigOption
}

func NewConfigx(opts ...ConfigOptions) (*Config, error) {
	configWithDefaults := defaultConfig()

	config := configWithDefaults

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		// *House as the argument
		opt(config)
	}

	// Sanitize
	if config.configDirEnv {
		if config.configDirEnvName != "" && config.envPrefix != "" {
			config.configDirEnvName = config.envPrefix + "_" + strings.ToUpper(config.configDirEnvName)
		}
	}

	err := config.initialize()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func defaultConfig() *Config {
	return &Config{
		tag:                    "config",
		envLoad:                true,
		envDotFileLoad:         true,
		envDotFilePath:         "./.env",
		envPrefix:              "",
		configFileLoad:         true,
		configFileName:         "config",
		configFileDir:          ".",
		configDirEnv:           true,
		configDirEnvName:       "CONFIG_DIR",
		configEnvExpand:        true,
		validateStopOnFirstErr: false,
		validateRecursive:      true,
		validateUsingTags:      true,
	}
}

func (c *Config) initialize() error {

	// Init Viper
	c.viper = viper.New()

	if c.envLoad {
		if c.envPrefix != "" {
			envPrefix := c.envPrefix
			c.viper.SetEnvPrefix(envPrefix)

			// Viper add the `prefix` + '_' to the Key *before* passing it to Key Replacer,causing the replacer to replace the '_' with '__' when it shouldn't.
			// by adding the Prefix to the replacer twice, this will let the replacer escapes the prefix as it scans through the string.
			c.viper.SetEnvKeyReplacer(strings.NewReplacer(envPrefix+"_", envPrefix+"_", "_", "__", ".", "_", "-", "_"))
		} else {
			c.viper.SetEnvKeyReplacer(strings.NewReplacer("_", "__", ".", "_", "-", "_"))
		}

		c.viper.AutomaticEnv()

		// load .env vars
		if _, err := os.Stat(c.envDotFilePath); err == nil || !os.IsNotExist(err) {
			err := godotenv.Load(c.envDotFilePath)
			if err != nil {
				return fmt.Errorf("error loading .env envvars from \"%s\": %s", c.envDotFilePath, err.Error())
			}
		}
	}

	if c.configFileLoad {
		// Config Name
		c.viper.SetConfigName(strings.ToLower(c.configFileName))

		// Config Dir Path
		configFileDir := c.configFileDir

		// Override with Env ?
		if c.configDirEnv {
			configDirEnvValue, isSet := os.LookupEnv(c.configDirEnvName)
			if isSet {
				configFileDir = configDirEnvValue
			}
		}

		c.viper.AddConfigPath(configFileDir + "/")
		c.viper.ConfigFileUsed()
	}

	// decoder config
	DefaultDecodeHookFuncs := []mapstructure.DecodeHookFunc{
		stringJSONArrayToSlice(),
		stringJSONObjToMap(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToIPHookFunc(),
	}

	if c.configEnvExpand {
		DefaultDecodeHookFuncs = append([]mapstructure.DecodeHookFunc{expandEnvVariablesWithDefaults()}, DefaultDecodeHookFuncs...)
	}

	c.decodeHook = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		DefaultDecodeHookFuncs...,
	))

	// Tag validator
	if c.validateUsingTags {
		c.validator = validator.New()
		// Get English Errors
		uni := ut.New(en.New(), en.New())
		c.validatorTrans, _ = uni.GetTranslator("en")
		_ = ens.RegisterDefaultTranslations(c.validator, c.validatorTrans)
	}

	return nil
}

type ConfigOptions func(*Config)

func LoadFromEnvironmentVariables(Enabled bool, EnvPrefix string) ConfigOptions {
	return func(h *Config) {
		h.envLoad = Enabled
		h.envPrefix = strings.ToUpper(EnvPrefix)
	}
}

func Tag(tag string) ConfigOptions {
	return func(h *Config) {
		h.tag = tag
	}
}

func LoadDotEnvFile(Enabled bool, envDotFilePath string) ConfigOptions {
	return func(h *Config) {
		h.envDotFileLoad = Enabled
		h.envDotFilePath = envDotFilePath
	}
}

func ExpandEnvironmentVariables(Enabled bool) ConfigOptions {
	return func(h *Config) {
		h.configEnvExpand = Enabled
	}
}

func Validate(validateStopOnFirstErr, validateRecursive, validateUsingTags bool) ConfigOptions {
	return func(h *Config) {
		h.validateStopOnFirstErr = validateStopOnFirstErr
		h.validateRecursive = validateRecursive
		h.validateUsingTags = validateUsingTags
	}
}

func LoadFromConfigFile(Enabled bool, fileName string, fileDirPath string, overrideDirWithEnv bool, configDirEnvName string) ConfigOptions {
	return func(h *Config) {
		h.configFileLoad = Enabled
		h.configFileName = fileName
		h.configFileDir = fileDirPath
		h.configDirEnv = overrideDirWithEnv
		h.configDirEnvName = configDirEnvName
	}
}
