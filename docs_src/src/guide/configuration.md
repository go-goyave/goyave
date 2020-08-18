---
meta:
  - name: "og:title"
    content: "Configuration - Goyave"
  - name: "twitter:title"
    content: "Configuration - Goyave"
  - name: "title"
    content: "Configuration - Goyave"
---

# Configuration

[[toc]]

## Introduction


The Goyave framework lets you configure its core and your application.
To configure your application, use the `config.json` file at your project's root. If you are using the template project, copy `config.example.json` to `config.json`. `config.json` should be ignored in your `.gitignore` file as it can differ from one developer to another. To avoid accidental commit or change to the default project's config, it is a good practice to ignore this file and define the project's default config in `config.example.json`.

If this config file misses some config entries, the default values will be used. Refer to the [configuration reference](#configuration-reference) to know more. 

All entries are **validated**. That means that the application will not start if you provided an invalid value in your config (for example if the specified port is not a number). That also means that a goroutine trying to change a config entry with the incorrect type will panic.  
Entries can be registered with a default value, their type and authorized values from any package. 

Configuration can be used concurrently safely.

The following JSON file is an example of default configuration:

```json
{
  "app": {
    "name": "goyave_template",
    "environment": "localhost",
    "debug": true,
    "defaultLanguage": "en-US"
  },
  "server": {
    "host": "127.0.0.1",
    "maintenance": false,
    "protocol": "http",
    "domain": "",
    "port": 8080,
    "httpsPort": 8081,
    "timeout": 10,
    "maxUploadSize": 10,
    "tls": {
      "cert": "/path/to/cert",
      "key": "/path/to/key"
    },
  },
  "database": {
    "connection": "mysql",
    "host": "127.0.0.1",
    "port": 3306,
    "name": "goyave",
    "username": "root",
    "password": "root",
    "options": "charset=utf8&parseTime=true&loc=Local",
    "maxOpenConnections": 20,
    "maxIdleConnections": 20,
    "maxLifetime": 300,
    "autoMigrate": false
  }
}
```

## Terminology

**Entry**: a configuration entry is a value accessible using a key.

**Registering an entry**: informs the framework that an entry with the given key is expected. Registering an entry allows to set a default value to be used if this entry is not provided in an app's configuration file, to enforce a certain type for this entry (for example if it needs to be an integer), and to set a list of allowed values.

**Category**: a category is represented by a JSON object in your configuration file, delimited by braces. Sub-categories are categories that are not at root level, for example: `server.tls` is a sub-category of the `server` category.

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

When the configuration is loaded, all default values are copied to the newly created map holding the configuration. Then, the configuration file is read, parsed and is applied over. This means that all entries from the file override the default ones. However, if an entry has a default value and the same entry is not present in the configuration file, then it is kept as it is. On the other hand, if an entry is present in the configuration file and not in the default values (meaning that this entry is not expected), a new entry will be registered.This new entry will be subsequently validated using the type of its initial value and have an empty slice as authorized values (meaning it can have any value of its type)

The following cases will raise errors when the configuration is being overridden:
- When the configuration file overrides an entry with a category
- When the configuration file overrides a category with an entry

#### config.Load

| Parameters | Return  |
|------------|---------|
|            | `error` |

**Example:**
``` go
config.Load()
```

### Getting a value

All entries are accessible using **dot-separated paths**. If you want to access the `name` entry in the `app` category, the key will be `app.name`.

#### config.Get

Get a generic config entry. 

Prefer using the `GetString`, `GetBool`, `GetInt` and `GetFloat` accessors. If you need a type not covered by those accessors, use `config.Get`. You may need to type-assert the returned value before using it. You can do so safely as the config values and types are validated.

Panics if the entry doesn't exist.

| Parameters   | Return                 |
|--------------|------------------------|
| `key string` | `interface{}` or panic |

**Example:**
``` go
config.Get("app.name") // "goyave"
```

#### config.GetString

Get a string config entry. Panics if the entry doesn't exist or is not a `string` or if it doesn't exist.

| Parameters   | Return            |
|--------------|-------------------|
| `key string` | `string` or panic |

**Example:**
``` go
config.GetString("server.protocol") // "http"
```

#### config.GetBool

Get a bool config entry. Panics if the entry doesn't exist or is not a `bool` or if it doesn't exist.

| Parameters   | Return          |
|--------------|-----------------|
| `key string` | `bool` or panic |

**Example:**
``` go
config.GetBool("app.debug") // true
```

#### config.GetInt

<p><Badge text="Since v3.0.0"/></p>

Get an int config entry. Panics if the entry doesn't exist or is not an `int` or if it doesn't exist.

| Parameters   | Return         |
|--------------|----------------|
| `key string` | `int` or panic |

**Example:**
``` go
config.GetInt("server.port") // 8080
```

#### config.GetFloat

<p><Badge text="Since v3.0.0"/></p>

Get a float config entry. Panics if the entry doesn't exist or is not a `float64` or if it doesn't exist.

| Parameters   | Return             |
|--------------|--------------------|
| `key string` | `float64` or panic |

**Example:**
``` go
config.GetInt("server.port") // 8080
```

#### config.Has

Check if a config entry exists.

| Parameters   | Return |
|--------------|--------|
| `key string` | `bool` |

**Example:**
``` go
config.Has("app.name") // true
```

### Setting a value

You can set a config value at runtime with the `config.Set(key, value)` function. Bear in mind that this change **temporary** and will be lost after your application restarts or if the config is reloaded. This function is mainly used for testing purposes. Values set using this function are still being validated: your application will panic and revert changes if the validation doesn't pass.

Use `nil` to unset a value.

- A category cannot be replaced with an entry.
- An entry cannot be replaced with a category.
- New categories can be created with they don't already exist.
- New entries can be created if they don't already exist. This new entry will be subsequently validated using the type of its initial value and have an empty slice as authorized values (meaning it can have any value of its type)

#### config.Set

| Parameters          | Return          |
|---------------------|-----------------|
| `key string`        | `void` or panic |
| `value interface{}` |                 |

**Example:**
``` go
config.Set("app.name", "my awesome app")
```

## Custom config entries

Configuration can be expanded. It is very likely that a plugin or a package you're developing is using some form of options. These options can be added to the configuration system so it is not needed to set them in the code or to make some wiring.

#### config.Register

Register a new config entry and its validation.

Each module should register its config entries in an `init()` function, even if they don't have a default value, in order to ensure they will be validated.

Each module should use its own category and use a name both expressive and unique to avoid collisions. For example, the `auth` package registers, among others, `auth.basic.username` and `auth.jwt.expiry`, thus creating a category for its package, and two subcategories for its features.

To register an entry without a default value (only specify how it will be validated), set `Entry.Value` to `nil`.

Panics if an entry already exists for this key and is not identical to the one passed as parameter of this function. On the other hand, if the entries are identical, no conflict is expected so the configuration is left in its current state.

| Parameters          | Return          |
|---------------------|-----------------|
| `key string`        | `void` or panic |
| `kind config.Entry` |                 |

**Example:**
``` go
func init() {
  config.Register("my-plugin.name", config.Entry{
    Value:            "default value",
    Type:             reflect.String,
    AuthorizedValues: []interface{}{},
  })
  
  // Without a default value (only validation)
  config.Register("my-plugin.protocol", config.Entry{
    Value:            nil,
    Type:             reflect.String,
    AuthorizedValues: []interface{}{"ftp", "sftp", "scp"},
  })
}
```

## Configuration reference

### App category

| Entry           | Type     | Accepted values | Default     | Note                                                                                                           |
|-----------------|----------|-----------------|-------------|----------------------------------------------------------------------------------------------------------------|
| name            | `string` | any             | "goyave"    |                                                                                                                |
| environment     | `string` | any             | "localhost" |                                                                                                                |
| debug           | `bool`   | `true`, `false` | `true`      | When activated, print stacktrace on error and sends error message in response. **Disable this in production!** |
| defaultLanguage | `string` | any             | "en-US"     | See the [Localization](./advanced/localization.html)                                                           |

### Server category

| Entry         | Type      | Accepted values | Default     | Note                                                                      |
|---------------|-----------|-----------------|-------------|---------------------------------------------------------------------------|
| host          | `string`  | any             | "127.0.0.1" |                                                                           |
| domain        | `string`  | any             | ""          | Used for URL generation Leave empty to use IP instead.                    |
| protocol      | `string`  | "http", "https" | "http"      | See the [HTTPS](#setting-up-https) section                                |
| port          | `int`     | any             | `8080`      |                                                                           |
| httpsPort     | `int`     | any             | `8081`      |                                                                           |
| timeout       | `int`     | any             | `10`        | Timeout in seconds                                                        |
| maxUploadSize | `float64` | any             | `10`        | Maximum size of the request, in MiB                                       |
| maintenance   | `bool`    | `true`, `false` | `false`     | If `true`, start the server in maintenance mode. (Always return HTTP 503) |

#### TLS sub-category

| Entry | Type     | Accepted values | Default | Note                  |
|-------|----------|-----------------|---------|-----------------------|
| cert  | `string` | any             | none    | Path to your TLS cert |
| key   | `string` | any             | none    | Path to your TLS key  |

### Database category

| Entry              | Type     | Accepted values                                 | Default                                 | Note                                                      |
|--------------------|----------|-------------------------------------------------|-----------------------------------------|-----------------------------------------------------------|
| connection         | `string` | "none", "mysql", "postgres", "sqlite3", "mssql" | "none"                                  | See the [Database](./basics/database.html) guide          |
| host               | `string` | any                                             | "127.0.0.1"                             |                                                           |
| port               | `int`    | any                                             | `3306`                                  |                                                           |
| name               | `string` | any                                             | "goyave"                                |                                                           |
| username           | `string` | any                                             | "root"                                  |                                                           |
| password           | `string` | any                                             | "root"                                  |                                                           |
| otions             | `string` | any                                             | "charset=utf8&parseTime=true&loc=Local" |                                                           |
| maxOpenConnections | `int`    | any                                             | `20`                                    |                                                           |
| maxIdleConnections | `int`    | any                                             | `20`                                    |                                                           |
| maxLifetime        | `int`    | any                                             | `300`                                   | The maximum time (in seconds) a connection may be reused. |
| autoMigrate        | `bool`   | `true`, `false`                                 | `false`                                 | When activated, migrate all registered models at startup  |


## Setting up HTTPS

Setting up HTTPS on your Goyave application is easy. First, turn `server.protocol` to `https` in the config. Then, add the `server.tls.cert` and `server.tls.key` entries in the config. These two entries represent respectively the path to your TLS certificate and your TLS key.

**Certbot example:**
``` json
{
    ...
    "server": {
      "protocol": "https",
      "tls": {
        "cert": "/etc/letsencrypt/live/mydomain.com/cert.pem",
        "key": "/etc/letsencrypt/live/mydomain.com/privkey.pem"
      }
    },
    ...
}
```

Restart your server and you should now be able to connect securely to your Goyave application! All HTTP requests will automatically be redirected to HTTPS.
