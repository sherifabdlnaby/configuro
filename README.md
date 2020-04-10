<p align="center">
<img width="637px" src="https://user-images.githubusercontent.com/16992394/78989442-e7c1ae80-7b33-11ea-98c6-1d37ed276a3b.png">
</p>
<h2 align="center">An opinionated configuration loading and validation framework focused towards Containerized and <a href="https://12factor.net/config">12-Factor</a> compliant applications</h2>
<h6 align="center">Read configurations from Environment Variables exclusively, falls back to config files supporting most common configuration formats. and Supporting Expanding and Validation</h4>
<p align="center">
   <a>
      <img src="https://img.shields.io/github/v/tag/sherifabdlnaby/configuro?label=release&amp;sort=semver">
    </a>
    <a>
      <img src="https://github.com/sherifabdlnaby/configuro/workflows/Build/badge.svg">
    </a>
    <a href='https://coveralls.io/github/sherifabdlnaby/configuro'><img src='https://img.shields.io/coveralls/github/sherifabdlnaby/configuro?logo=codecov&logoColor=white' alt='Coverage Status' /></a>

   <a>
      <img src="https://img.shields.io/badge/Go-%3E=v1.13-blue?style=flat&logo=go" alt="Go Version">
   </a>

   <a href="https://goreportcard.com/report/github.com/sherifabdlnaby/configuro">
      <img src="https://goreportcard.com/badge/github.com/sherifabdlnaby/configuro" alt="Go Report">
   </a>
   <a href="https://github.com/sherifabdlnaby/configuro/issues">
        <img src="https://img.shields.io/github/issues/sherifabdlnaby/configuro.svg" alt="GitHub issues">
   </a>
   <a href="https://raw.githubusercontent.com/sherifabdlnaby/configuro/blob/master/LICENSE">
      <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="GitHub license">
   </a>
</p>

# Introduction

Configuro is an opinionated configuration loading and validation library with not very much to configure, ready to be used without much hassle and gluing together libraries.

It provides the set of features you would want to implement 12-Factor compliant config and be container ready.
**It has one defined strategy to load configurations** while only exposing few options for simplicity.

### Configuration Loading Strategy Briefly

1. Application has **a single custom configuration struct** that *Configuro* will `unmarshall` read configuration into.
2. The configuration is loaded from a file `config`, it can be in `Yaml`, `Json`, `Toml` `ini` or `hcl` format.
3. Configuration values can be loaded/overloaded **by Environment Variables**.
    - The key `database.host` can be expressed in environment variables as `DATABASE_HOST`; if set, it will take precedence over config file's value.
    - You can require all Environment Variables **to be prefixed** with your own string. e.g `database.host` => `APPNAME_DATABASE_HOST`.
    - If the key itself contains `_` or `-` then replace them with `__` in the Environment Variable.
    - You can express **Maps** and **Lists** in Environment Variables by JSON encoding them. (e.g `CONFIG: {"a":123, "b": "abc"}`)
    - You can provide a `.env` file to load environment variables that are not set by the OS.
4. Configuration Values can have ${ENV} expression that will be expanded at loading time.
    - Example `host: www.example.com:%{PORT|3306}` with `3306` being the default value if env ${PORT} is not set).
5. Configuro can validate structs recursively using [Validation Tags](https://godoc.org/github.com/go-playground/validator), or by Implementing `Validatable` Interface `Validate() error`

#### Notes
- 📣 You can give up the config file option entirely and rely exclusively on Environment Variables.


-------------------------------------------------------------------------

# Install
``` bash
go get github.com/sherifabdlnaby/configuro
```
``` go
import "github.com/sherifabdlnaby/configuro"
```

# Usage

### 1. Start with Defining Your Config Struct
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

### 3. Loading from Configuration Files

- Upon setting up you will declare the config `filename` and its `directory` without needing to specify the extension or format, Configuro will look if a file with supported format exists and attempt loading from it.
    - Default `filename`  => "config"
    - Default `directory` => "." (current directory)
    - So, by default, any of `config.yml`, `config.json`, `config.toml`, etc are looked up in the `directory` to load config.
- Supported formats are `Yaml`, `Json`, `Toml` `ini` or `hcl`.
- Config file directory can be overloaded with a defined Environment Variable.
    - Default: `PREFIX_CONFIG_DIR`.

The above settings can be changed upon constructing the configuro object via passing these options.
```go
    configuro.LoadFromConfigFile(enabled, fileName, fileDirPath)
    configuro.OverloadConfigPathWithEnv(enabled, envVarName)
```

### 4. Loading from Environment Variables

- Values found in Environment Variables take precedence over values found in config file.
- The key `database.host` can be expressed in environment variables as `DATABASE_HOST`.
- If the key itself contains `_` or `-` then replace them with `__` in the Environment Variable.
- You can require all Environment Variables **to be prefixed** with your own string. e.g `database.host` => `APPNAME_DATABASE_HOST`.
- You can express **Maps** and **Lists** in Environment Variables by JSON encoding them. (e.g `CONFIG: {"a":123, "b": "abc"}`)
- You can provide a `.env` file to load environment variables that are not set by the OS. (notice that .env is loaded globally in the application scope)
- Loading from Environment is **Enabled by default with no prefix** when constructing Configuro.

The above settings can be changed upon constructing the configuro object via passing these options.
```go
    configuro.LoadFromEnvironmentVariables(enabled, prefix)
    configuro.LoadDotEnvFile(enabled, envDotFilePath)
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
    configuro.ExpandEnvironmentVariables(enabled)
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
- Validation returns an error of type [multierror](https://github.com/hashicorp/go-multierror) specifying each field that error-ed.
- It can be configured to not recursively validate types with `Validatable` Interface. (default: recursively)
- It can be configured to stop at the first error. (default: false)
- It can be configured to not use Validation Tags. (default: false)
The above settings can be changed upon constructing the configuro object via passing these options.
```go
    configuro.Validate(validateStopOnFirstErr, validateRecursive, validateUsingTags)
```

### 7. Miscellaneous

- `config` and `validate` tag can be renamed using `configuro.Tag(structTag, validateTag)` construction option.

# License
[MIT License](https://raw.githubusercontent.com/sherifabdlnaby/configuro/master/LICENSE)
Copyright (c) 2020 Sherif Abdel-Naby

# Contribution

PR(s) are Open and Welcomed.
