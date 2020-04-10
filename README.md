<p align="center">
<img width="637px" src="https://user-images.githubusercontent.com/16992394/78989442-e7c1ae80-7b33-11ea-98c6-1d37ed276a3b.png">
</p>
<h2 align="center">An opinionated configuration loading and validation framework focused towards Containerized and <a href="https://12factor.net/config">12-Factor</a> compliant applications</h2>
<p align="center">
   <a>
      <img src="https://img.shields.io/github/v/tag/sherifabdlnaby/configuro?label=release&amp;sort=semver">
    </a>
    <a>
      <img src="https://github.com/sherifabdlnaby/configuro/workflows/Test/badge.svg">
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

It provide the set of features you would want to implement 12-Factor config and be container ready.
It has one defined strategy to load configurations and only exposing few options for simplicity.

### Configuration Loading Stratgey

1. Application has single custom configuration struct that Configuro will `unmarshall` read configuration into.
2. Configuration is loaded from a file `config`, it can be in `Yaml`, `Json`, `Toml` `ini` or `hcl` format.
3. Configuration Keys can be overloaded **by Enviroment Variables**.
    - Example A value with key `database.host` can be expressed in environment variables as follows `APPNAME_DATABASE_HOST`, and if set, it will take precedence over config file's value.
4. `.env` file is read on to load Enviroment Variables that is not set by the OS.
5. Configuration Values can have ${ENV} expression that will be expanded at loading time.
    - Example `host: www.example.com:%{PORT|8080}` with `8080` being the default value if env ${PORT} is not set).
6. Configuro can validate structs recursivly using [Validation Tags](https://godoc.org/github.com/go-playground/validator), or by Implementing `Validatable` Interface `Validate() error`

#### Notes
- 📣 You can give up the config file option entirely and rely exclusively on Environment Variables.

# Usage

``` bash
go get github.com/sherifabdlnaby/configuro
```
``` go
import "github.com/sherifabdlnaby/configuro"
```

# License
[MIT License](https://raw.githubusercontent.com/sherifabdlnaby/configuro/master/LICENSE)
Copyright (c) 2020 Sherif Abdel-Naby

# Contribution

PR(s) are Open and Welcomed.
