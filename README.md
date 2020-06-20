<p align="center">
<br/>
<br/>
<img width="637px" src="https://user-images.githubusercontent.com/16992394/78989442-e7c1ae80-7b33-11ea-98c6-1d37ed276a3b.png">
</p>
<h3 align="center">Opinionated configuration loading framework for Containerized and <a href="https://12factor.net/config">12-Factor</a> compliant applications.</h3>
<h6 align="center">Read configurations from Environment Variables, and/or Configuration Files. With support to Environment Variables Expanding and Validation Methods.</h4>
<p align="center">
   <a href="https://github.com/avelino/awesome-go#configuration">
      <img src="https://awesome.re/mentioned-badge.svg" alt="Mentioned in Awesome Go">
   </a>
   <a href="http://godoc.org/github.com/sherifabdlnaby/configuro">
      <img src="https://godoc.org/github.com/sherifabdlnaby/configuro?status.svg" alt="Go Doc">
   </a>
   <a>
      <img src="https://img.shields.io/github/v/tag/sherifabdlnaby/configuro?label=release&amp;sort=semver">
    </a>
   <a>
      <img src="https://img.shields.io/badge/Go-%3E=v1.11-blue?style=flat&logo=go" alt="Go Version">
   </a>
    <a>
      <img src="https://github.com/sherifabdlnaby/configuro/workflows/Build/badge.svg">
    </a>
    <a href='https://coveralls.io/github/sherifabdlnaby/configuro'><img src='https://img.shields.io/coveralls/github/sherifabdlnaby/configuro?logo=codecov&logoColor=white' alt='Coverage Status' /></a>
   <a href="https://goreportcard.com/report/github.com/sherifabdlnaby/configuro">
      <img src="https://goreportcard.com/badge/github.com/sherifabdlnaby/configuro" alt="Go Report">
   </a>
   <a href="https://raw.githubusercontent.com/sherifabdlnaby/configuro/blob/master/LICENSE">
      <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="GitHub license">
   </a>
</p>

# Introduction

Configuro is an opinionated configuration loading and validation framework with not very much to configure. It *defines* a method for loading configurations without so many options for a straight-forward and simple code.

The method defined by Configuro allow you to implement [12-Factor's Config](https://12factor.net/config) and it mimics how many mature applications do configurations (e.g Elastic, Neo4J, etc); and is also fit for containerized applications.

------------
**With Only with two lines of code, and zero setting up** you get loading configuration from Config File, Overwrite values /or rely exclusively on Environment Variables, Expanding values using ${ENV} expressions, and validation tags.

# Loading Configuration
<p align="center">
<img width="770" src="https://user-images.githubusercontent.com/16992394/85208005-4b830780-b32d-11ea-9b41-83a3c87704e3.png">
</p>

### 1. Define Application **configurations** in a struct
- Which *Configuro* will `Load()` the read configuration into.

### 2. Setting Configuration by **Environment Variables**.
- Value for key `database.password` can be set by setting `CONFIG_DATABASE_PASSWORD`. (`CONFIG_` default prefix can be changed)
- If the key itself contains `_` then replace with `__` in the Environment Variable.
- You can express **Maps** and **Lists** in Environment Variables by JSON encoding them. (e.g `CONFIG: {"a":123, "b": "abc"}`)
- You can provide a `.env` file to load environment variables that are not set by the OS.

### 3. Setting Configuration by Configuration File.
- Defaults to `config.yml`; name and extension can be configured.
- Supported extensions are `.yml`, `.yaml`, `.json`, and `.toml`.

### 4. Support Environment Variables Expanding.
- Configuration Values can have ${ENV|default} expression that will be expanded at loading time.
- Example `host: www.example.com:%{PORT|3306}` with `3306` being the default value if env ${PORT} is not set).

### 5. Validate Loaded Values
- Configuro can validate structs recursively using [Validation Tags](https://godoc.org/github.com/go-playground/validator).
- By Implementing `Validatable` Interface `Validate() error`.

#### Notes
- 📣 Values' precedence is `OS EnvVar` > `.env EnvVars` > `Config File` > `Value set in Struct before loading.`


-------------------------------------------------------------------------

# Install
``` bash
go get github.com/sherifabdlnaby/configuro
```
``` go
import "github.com/sherifabdlnaby/configuro"
```

# Usage

### 1. Define You Config Struct.
This is the struct you're going to use to retrieve your config values in-code.
```go
type Config struct {
    Database struct {
        Host     string
        Port     int
    }
    Logging struct {
        Level  string
        LineFormat string `config:"line_format"`
    }
}
```

- Nested fields accessed with `.` (e.g `Database.Host` )
- Use `config` tag to change field name if you want it to be different from the Struct field name.
- All [mapstructure](https://github.com/mitchellh/mapstructure) tags apply to `config` tag for unmarshalling.
- Fields must be public to be accessible by Configuro.

### 2. Create and Configure the `Configuro.Config` object.

```go
    // Create Configuro Object with supplied options (explained below)
    config, err := configuro.NewConfig( opts ...configuro.ConfigOption )

    // Create our Config Struct
    configStruct := &Config{ /*put defaults config here*/ }

    // Load values in our struct
    err = config.Load(configStruct)
```

- Create Configuro Object Passing to the constructor `opts ...configuro.ConfigOption` which is explained in the below sections.
- This should happen as early as possible in the application.

### 3. Loading from Environment Variables

- Values found in Environment Variables take precedence over values found in config file.
- The key `database.host` can be expressed in environment variables as `CONFIG_DATABASE_HOST`.
- If the key itself contains `_`  then replace them with `__` in the Environment Variable.
- `CONFIG_` prefix can be configured.
- You can express **Maps** and **Lists** in Environment Variables by JSON encoding them. (e.g `CONFIG: {"a":123, "b": "abc"}`)
- You can provide a `.env` file to load environment variables that are not set by the OS. (notice that .env is loaded globally in the application scope)

The above settings can be changed upon constructing the configuro object via passing these options.
```go
    configuro.WithLoadFromEnvVars(EnvPrefix string)  // Enable Env loading and set Prefix.
    configuro.WithoutLoadFromEnvVars()               // Disable Env Loading Entirely
    configuro.WithLoadDotEnv(envDotFilePath string)  // Enable loading .env into Environment Variables
    configuro.WithoutLoadDotEnv()                    // Disable loading .env
```

### 4. Loading from Configuration Files
- Upon setting up you will declare the config `filepath`.
    - Default `filename`  => "config.yml"
- Supported formats are `Yaml`, `Json`, and `Toml`.
- Config file directory can be overloaded with a defined Environment Variable.
    - Default: `CONFIG_DIR`.
- If file **was not found** Configuro won't raise an error unless configured too. This is you can rely 100% on Environment Variables.

The above settings can be changed upon constructing the configuro object via passing these options.
```go
    configuro.WithLoadFromConfigFile(Filepath string, ErrIfFileNotFound bool)    // Enable Config File Load
    configuro.WithoutLoadFromConfigFile()                                        // Disable Config File Load
    configuro.WithEnvConfigPathOverload(configFilepathENV string)                // Enable Overloading Path with ENV var.
    configuro.WithoutEnvConfigPathOverload()                                     // Disable Overloading Path with ENV var.
```

### 5. Expanding Environment Variables in Config

- `${ENV}` and `${ENV|default}` expressions are evaluated and expanded if the Environment Variable is set or with the default value if defined, otherwise it leaves it as it is.
```
config:
    database:
        host: xyz:${PORT:3306}
        username: admin
        password: ${PASSWORD}
```

The above settings can be changed upon constructing the configuro object via passing these options.
```go
    configuro.WithExpandEnvVars()       // Enable Expanding
    configuro.WithoutExpandEnvVars()    // Disable Expanding
```

### 6. Validate Struct

```go
    err := config.Validate(configStruct)
    if err != nil {
        return err
    }
````

- Configuro Validate Config Structs using two methods.
    1. Using [Validation Tags](https://godoc.org/github.com/go-playground/validator) for quick validations.
    2. Using `Validatable` Interface that will be called on any type that implements it recursively, also on each element of a Map or a Slice.
- Validation returns an error of type configuro.ErrValidationErrors if more than error occurred.
- It can be configured to not recursively validate types with `Validatable` Interface. (default: recursively)
- It can be configured to stop at the first error. (default: false)
- It can be configured to not use Validation Tags. (default: false)
The above settings can be changed upon constructing the configuro object via passing these options.
```go
    configuro.WithValidateByTags()
    configuro.WithoutValidateByTags()
    configuro.WithValidateByFunc(stopOnFirstErr bool, recursive bool)
    configuro.WithoutValidateByFunc()
```

### 7. Miscellaneous

- `config` and `validate` tag can be renamed using `configuro.Tag(structTag, validateTag)` construction option.

# Built on top of
- [spf13/viper](https://github.com/spf13/viper)
- [mitchellh/mapstructure](github.com/mitchellh/mapstructure)
- [go-playground/validator](https://github.com/go-playground/validator)
- [joho/godotenv](https://github.com/joho/godotenv)

# License
[MIT License](https://raw.githubusercontent.com/sherifabdlnaby/configuro/master/LICENSE)
Copyright (c) 2020 Sherif Abdel-Naby

# Contribution

PR(s) are Open and Welcomed.
