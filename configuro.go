package configuro

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	ens "github.com/go-playground/validator/translations/en"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/go-playground/validator.v9"
)

//Config Loads and Validate Arbitrary structs based on options (set at constructing)
type Config struct {
	envLoad                 bool
	envPrefix               string
	envDotFileLoad          bool
	envDotFilePath          string
	configFileLoad          bool
	configFileErrIfNotFound bool
	configFileName          string
	configFileDir           string
	configDirEnv            bool
	configDirEnvName        string
	configEnvExpand         bool
	validateStopOnFirstErr  bool
	validateRecursive       bool
	validateUsingTags       bool
	validateTag             string
	tag                     string
	viper                   *viper.Viper
	validator               *validator.Validate
	validatorTrans          ut.Translator
	decodeHook              viper.DecoderConfigOption
}

//NewConfig Create config Loader/Validator according to options.
func NewConfig(opts ...ConfigOptions) (*Config, error) {
	var err error

	configWithDefaults := defaultConfig()
	config := configWithDefaults

	// Loop through each option
	for _, opt := range opts {
		// Call the option giving the instantiated
		// *House as the argument
		err = opt(config)
		if err != nil {
			return nil, err
		}

	}

	err = config.initialize()
	if err != nil {
		return nil, err
	}

	return config, nil
}

func defaultConfig() *Config {
	return &Config{
		tag:                    "config",
		validateTag:            "validate",
		envLoad:                true,
		envDotFileLoad:         true,
		envDotFilePath:         "./.env",
		envPrefix:              "CONFIG",
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
		c.enableEnvLoad()
	}

	if c.configFileLoad {
		c.enableConfigFileLoad()
	}

	// decoder config
	c.addDecoderConfig()

	// Tag validator
	if c.validateUsingTags {
		c.enableValidateUsingTag()
	}

	return nil
}

// ---------------------------------------------------------------------------------------------------------------------

//ConfigOptions Modify Config Options Accordingly
type ConfigOptions func(*Config) error

//LoadFromEnvironmentVariables Load Configuration from Environment Variables if they're set.
// 	- Prefix Require Environment Variables to prefixed with the set prefix (All CAPS)
// 	- For Nested fields replace `.` with `_` and if key itself has any `_` or `-` replace with `__` (e.g `config.host` to be `CONFIG_HOST`)
//	- Arrays can be declared in environment variables using
//		1. comma separated list.
//		2. json encoded array in a string.
//	- Maps and objects can be declared in environment using a json encoded object in a string.
func LoadFromEnvironmentVariables(Enabled bool, EnvPrefix string) ConfigOptions {
	return func(h *Config) error {
		h.envLoad = Enabled
		if EnvPrefix == "" {
			return fmt.Errorf("env prefix must be declared")
		}
		h.envPrefix = strings.ToUpper(EnvPrefix)
		return nil
	}
}

//Tag Change default tag.
func Tag(structTag, validateTag string) ConfigOptions {
	return func(h *Config) error {
		h.tag = structTag
		h.validateTag = validateTag
		return nil
	}
}

//LoadDotEnvFile Allow loading .env file (notice that this is application global not to this config instance only)
func LoadDotEnvFile(Enabled bool, envDotFilePath string) ConfigOptions {
	return func(h *Config) error {
		h.envDotFileLoad = Enabled
		h.envDotFilePath = envDotFilePath
		return nil
	}
}

//ExpandEnvironmentVariables Expand config values with ${ENVVAR} with the value of ENVVAR in environment variables.
// You can set default if ENVVAR is not set using the following ${ENVVAR|defaultValue}
func ExpandEnvironmentVariables(Enabled bool) ConfigOptions {
	return func(h *Config) error {
		h.configEnvExpand = Enabled
		return nil
	}
}

//Validate Control Validate function behavior.
func Validate(validateStopOnFirstErr, validateRecursive, validateUsingTags bool) ConfigOptions {
	return func(h *Config) error {
		h.validateStopOnFirstErr = validateStopOnFirstErr
		h.validateRecursive = validateRecursive
		h.validateUsingTags = validateUsingTags
		return nil
	}
}

//LoadFromConfigFile Load Config from file (notice that file doesn't have an extension as any file with supported extension should work)
func LoadFromConfigFile(Enabled bool, fileName string, fileDirPath string) ConfigOptions {
	return func(h *Config) error {
		h.configFileLoad = Enabled
		h.configFileName = fileName
		h.configFileDir = fileDirPath + "/"
		return nil
	}
}

//OverloadConfigPathWithEnv Allow to override Config Dir Path with an Env Variable
func OverloadConfigPathWithEnv(overrideDirWithEnv bool, configDirEnvName string) ConfigOptions {
	return func(h *Config) error {
		h.configDirEnv = overrideDirWithEnv
		h.configDirEnvName = strings.ToUpper(configDirEnvName)
		return nil
	}
}

func (c *Config) enableValidateUsingTag() {
	c.validator = validator.New()
	c.validator.SetTagName(c.validateTag)
	// Get English Errors
	uni := ut.New(en.New(), en.New())
	c.validatorTrans, _ = uni.GetTranslator("en")
	_ = ens.RegisterDefaultTranslations(c.validator, c.validatorTrans)
}

func (c *Config) addDecoderConfig() {
	DefaultDecodeHookFuncs := []mapstructure.DecodeHookFunc{
		stringJSONArrayToSlice(),
		stringJSONObjToMap(),
		stringJSONObjToStruct(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToIPHookFunc(),
	}
	if c.configEnvExpand {
		DefaultDecodeHookFuncs = append([]mapstructure.DecodeHookFunc{expandEnvVariablesWithDefaults()}, DefaultDecodeHookFuncs...)
	}
	c.decodeHook = viper.DecodeHook(mapstructure.ComposeDecodeHookFunc(
		DefaultDecodeHookFuncs...,
	))
}

func (c *Config) enableConfigFileLoad() {
	// Config Name
	c.viper.SetConfigName(c.configFileName)
	// Config Dir Path
	configFileDir := c.configFileDir
	// Override with Nested ?
	//TODO make this after dot env.
	if c.configDirEnv {
		configDirEnvValue, isSet := os.LookupEnv(c.configDirEnvName)
		if isSet {
			configFileDir = configDirEnvValue
		}
	}
	c.viper.AddConfigPath(configFileDir + "/")
	c.viper.ConfigFileUsed()
}

func (c *Config) enableEnvLoad() {
	c.viper.SetEnvPrefix(c.envPrefix)
	// Viper add the `prefix` + '_' to the Key *before* passing it to Key Replacer,causing the replacer to replace the '_' with '__' when it shouldn't.
	// by adding the Prefix to the replacer twice, this will let the replacer escapes the prefix as it scans through the string.
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(c.envPrefix+"_", c.envPrefix+"_", "_", "__", ".", "_"))
	c.viper.AutomaticEnv()
}
