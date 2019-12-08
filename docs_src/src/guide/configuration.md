# Configuration

[[toc]]

## Introduction

The Goyave framework lets you configure its core and your application.
To configure your application, use the `config.json` file at your project's root. If you are using the template project, copy `config.example.json` to `config.json`. `config.json` should be ignored in your `.gitignore` file as it can differ from one developer to another. To avoid accidental commit or change to the default project's config, it is a good practice to ignore this file and define the project's default config in `config.example.json`.

If this config file misses some config entries, the default values will be used. All values from the framework's core are **validated**. That means that the application will not start if you provided an invalid value in your config (For example if the specified port is not a number).  
See the [configuration reference](#configuration-reference) for more details.

## Environment configuration

Most projects need different configuration values based on the environment. For example, you won't connect to the same database if you're in local development, in a testing environment inside continuous integration pipelines, or in production. Goyave supports multiple configurations and will pick the appropriate one depending on the environment variable `GOYAVE_ENV`.

Since `v2.0.0`, you can use custom environments.

| GOYAVE_ENV                    | Config file              |
|-------------------------------|--------------------------|
| test                          | `config.test.json`       |
| production                    | `config.production.json` |
| *custom_env*                  | `config.custom_env.json` |
| local / localhost / *not set* | `config.json`            |

## Using the configuration

Before being able to use the config, import the config package:
``` go
import "github.com/System-Glitch/goyave/v2/config"
```

The configuration is loaded automatically when the server starts, but you can reload it manually if needed.

#### config.Load

| Parameters | Return  |
|------------|---------|
|            | `error` |

**Example:**
``` go
config.Load()
```

### Getting a value

#### config.Get

Get a generic config entry. You may need to type-assert it before being able to use it. You can do so safely as the config values and types are validated. Panics if the entry doesn't exist.

| Parameters   | Return                 |
|--------------|------------------------|
| `key string` | `interface{}` or panic |

**Example:**
``` go
config.Get("appName") // "goyave"
```

#### config.GetString

Get a string config entry. Panics if the entry doesn't exist or is not a string.

| Parameters   | Return            |
|--------------|-------------------|
| `key string` | `string` or panic |

**Example:**
``` go
config.GetString("protocol") // "http"
```

#### config.GetBool

Get a bool config entry. Panics if the entry doesn't exist or is not a bool.

| Parameters   | Return          |
|--------------|-----------------|
| `key string` | `bool` or panic |

**Example:**
``` go
config.GetBool("debug") // true
```

### Setting a value

You can set a config value at runtime with the `config.Set(key, value)` function. Bear in mind that this change **temporary** and will be lost after your application restarts or if the config is reloaded. This function is mainly used for testing purposes. Values set using this function are still being validated, and your application will panic if the validation doesn't pass.

#### config.Set

| Parameters          | Return          |
|---------------------|-----------------|
| `key string`        | `void` or panic |
| `value interface{}` |                 |

**Example:**
``` go
config.Set("appName", "my awesome app")
```

## Custom config entries

The core of the framework contains default values covering most cases, but you can still add custom entries for your most specific needs to the config simply by appending a property to your application's config file. The custom properties are not validated, so you can use the data type you want.

## Configuration reference

| Entry                | Type      | Accepted values                                 | Default                                 | Note                                                                                                           |
|----------------------|-----------|-------------------------------------------------|-----------------------------------------|----------------------------------------------------------------------------------------------------------------|
| appName              | `string`  | any                                             | "goyave"                                |                                                                                                                |
| environment          | `string`  | any                                             | "localhost"                             |                                                                                                                |
| maintenance          | `bool`    | `true`, `false`                                 | `false`                                 | If `true`, start the server in maintenance mode. (Always return HTTP 503)                                      |
| host                 | `string`  | any                                             | "127.0.0.1"                             |                                                                                                                |
| port                 | `float64` | any                                             | `8080`                                  |                                                                                                                |
| httpsPort            | `float64` | any                                             | `8081`                                  |                                                                                                                |
| protocol             | `string`  | "http", "https"                                 | "http"                                  | See the [HTTPS](#setting-up-https) section                                                                     |
| tlsCert              | `string`  | any                                             | none                                    | Path to your TLS cert                                                                                          |
| tlsKey               | `string`  | any                                             | none                                    | Path to your TLS key                                                                                           |
| debug                | `bool`    | `true`, `false`                                 | `true`                                  | When activated, print stacktrace on error and sends error message in response. **Disable this in production!** |
| timeout              | `float64` | any                                             | `10`                                    | Timeout in seconds                                                                                             |
| maxUploadSize        | `float64` | any                                             | `10`                                    | Max **in-memory** files sent in the request, in MiB                                                            |
| defaultLanguage      | `string`  | any                                             | "en-US"                                 | See the [Localization](./advanced/localization.html) guide                                                          |
| dbConnection         | `string`  | "none", "mysql", "postgres", "sqlite3", "mssql" | "none"                                  | See the [Database](./basics/database.html) guide                                                                    |
| dbHost               | `string`  | any                                             | "127.0.0.1"                             |                                                                                                                |
| dbPort               | `float64` | any                                             | `3306`                                  |                                                                                                                |
| dbName               | `string`  | any                                             | "goyave"                                |                                                                                                                |
| dbUsername           | `string`  | any                                             | "root"                                  |                                                                                                                |
| dbPassword           | `string`  | any                                             | "root"                                  |                                                                                                                |
| dbOptions            | `string`  | any                                             | "charset=utf8&parseTime=true&loc=Local" |                                                                                                                |
| dbMaxOpenConnections | `float64` | any                                             | `100`                                   |                                                                                                                |
| dbMaxIdleConnections | `float64` | any                                             | `20`                                    |                                                                                                                |
| dbAutoMigrate        | `bool`    | `true`, `false`                                 | `false`                                 | When activated, migrate all registered models at startup                                                       |

::: tip Note
Numeric values are parsed as `float64` even if they are supposed to be integers so it covers the potential use-case of floats in the config.
:::

## Setting up HTTPS

Setting up HTTPS on your Goyave application is easy. First, turn `protocol` to `https` in the config. Then, add the `tlsCert` and `tlsKey` entries in the config. These two entries represent respectively the path to your TLS certificate and your TLS key.

**Certbot example:**
``` json
{
    ...
    "protocol": "https",
    "tlsCert": "/etc/letsencrypt/live/mydomain.com/cert.pem",
    "tlsKey": "/etc/letsencrypt/live/mydomain.com/privkey.pem",
    ...
}
```

Restart your server and you should now be able to connect securely to your Goyave application! All HTTP requests will automatically be redirected to HTTPS.
