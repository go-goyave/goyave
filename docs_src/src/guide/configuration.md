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
    }
  },
  "database": {
    "connection": "mysql",
    "host": "127.0.0.1",
    "port": 3306,
    "name": "goyave",
    "username": "root",
    "password": "root",
    "options": "charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local",
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
import "github.com/System-Glitch/goyave/v3/config"
```

The configuration is loaded automatically when the server starts, but you can reload it manually if needed.

When the configuration is loaded, all default values are copied to the newly created map holding the configuration. Then, the configuration file is read, parsed and is applied over. This means that all entries from the file override the default ones. However, if an entry has a default value and the same entry is not present in the configuration file, then it is kept as it is. On the other hand, if an entry is present in the configuration file and not in the default values (meaning that this entry is not expected), a new entry will be registered. This new entry will be subsequently validated using the type of its initial value and have an empty slice as authorized values (meaning it can have any value of its type).

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

#### config.LoadFrom

You may need to load a configuration file from a custom path instead of using the standard one. `LoadFrom` lets you load a config file from the given path.

| Parameters    | Return  |
|---------------|---------|
| `path string` | `error` |

**Example:** load the config from the path given through a flag
``` go
import (
  "flag"
  "os"

  "github.com/System-Glitch/goyave/v3"
  "github.com/System-Glitch/goyave/v3/config"
  
  //...
)

func handleFlags() {
	flag.Usage = func() {
		goyave.ErrLogger.Println("usage: " + os.Args[0] + " -config=[config]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	flag.String("config", "", "JSON config file")
	flag.Parse()

	configDir := flag.Lookup("config")
	path := configDir.Value.String()
	if path != configDir.DefValue {
		if err := config.LoadFrom(path); err != nil {
			goyave.ErrLogger.Println(err)
			os.Exit(goyave.ExitInvalidConfig)
		}
	}
}

func main() {
	handleFlags()

	if err := goyave.Start(route.Register); err != nil {
		os.Exit(err.(*goyave.Error).ExitCode)
	}
}
```

### Using environment variables

<p><Badge text="Since v3.0.0"/></p>

You can use environment variables in your configuration file. Environment variables are identified by the following syntax: `${VARIABLE_NAME}`.

```json
{
  "database": {
    "host": "${DB_HOST}"
  }
}
```

**Note:** *This syntax is strict. If the string doesn't start with `${` or doesn't end with `}`, it will not be considered an environment variable.*

`int`, `float64` and `bool` values are supported. If the configuration entry is expected to be of one of these types, the content of the environment variable will be automatically converted. If the conversion fails, a configuration loading error will be returned.

If an environment variable mentioned in a configuration file is not set, the configuration validation will not pass. Environment variables are not supported inside slices.

### Getting a value

All entries are accessible using **dot-separated paths**. If you want to access the `name` entry in the `app` category, the key will be `app.name`.

::: table
[Get](#config-get)
[GetString](#config-getstring)
[GetBool](#config-getbool)
[GetInt](#config-getint)
[GetFloat](#config-getfloat)
[GetStringSlice](#config-getstringslice)
[GetBoolSlice](#config-getboolslice)
[GetIntSlice](#config-getintslice)
[GetFloatSlice](#config-getfloatslice)
[Has](#config-has)
:::

#### config.Get

Get a generic config entry. 

Prefer using the `GetString`, `GetBool`, `GetInt`, `GetFloat`, ... accessors. If you need a type not covered by those accessors, use `config.Get`. You may need to type-assert the returned value before using it. You can do so safely as the config values and types are validated.

Panics if the entry doesn't exist.

| Parameters   | Return                 |
|--------------|------------------------|
| `key string` | `interface{}` or panic |

**Example:**
``` go
config.Get("app.name") // "goyave"
```

#### config.GetString

Get a string config entry. Panics if the entry doesn't exist or is not a `string`.

| Parameters   | Return            |
|--------------|-------------------|
| `key string` | `string` or panic |

**Example:**
``` go
config.GetString("server.protocol") // "http"
```

#### config.GetBool

Get a bool config entry. Panics if the entry doesn't exist or is not a `bool`.

| Parameters   | Return          |
|--------------|-----------------|
| `key string` | `bool` or panic |

**Example:**
``` go
config.GetBool("app.debug") // true
```

#### config.GetInt

<p><Badge text="Since v3.0.0"/></p>

Get an int config entry. Panics if the entry doesn't exist or is not an `int`.

| Parameters   | Return         |
|--------------|----------------|
| `key string` | `int` or panic |

**Example:**
``` go
config.GetInt("server.port") // 8080
```

#### config.GetFloat

<p><Badge text="Since v3.0.0"/></p>

Get a float config entry. Panics if the entry doesn't exist or is not a `float64`.

| Parameters   | Return             |
|--------------|--------------------|
| `key string` | `float64` or panic |

**Example:**
``` go
config.GetInt("server.port") // 8080
```

#### config.GetStringSlice

<p><Badge text="Since v3.0.0"/></p>

Get a string slice config entry. Panics if the entry doesn't exist or is not a `[]string`.

| Parameters   | Return              |
|--------------|---------------------|
| `key string` | `[]string` or panic |

**Example:**
``` go
config.GetStringSlice("stringSlice") // [val1 val2]
```

#### config.GetBoolSlice

<p><Badge text="Since v3.0.0"/></p>

Get a bool slice config entry. Panics if the entry doesn't exist or is not a `[]bool`.

| Parameters   | Return            |
|--------------|-------------------|
| `key string` | `[]bool` or panic |

**Example:**
``` go
config.GetBoolSlice("boolSlice") // [true false]
```

#### config.GetIntSlice

<p><Badge text="Since v3.0.0"/></p>

Get an int slice config entry. Panics if the entry doesn't exist or is not a `[]int`.

| Parameters   | Return           |
|--------------|------------------|
| `key string` | `[]int` or panic |

**Example:**
``` go
config.GetIntSlice("intSlice") // [3 5]
```

#### config.GetFloatSlice

<p><Badge text="Since v3.0.0"/></p>

Get a float64 slice config entry. Panics if the entry doesn't exist or is not a `[]float64`.

| Parameters   | Return               |
|--------------|----------------------|
| `key string` | `[]float64` or panic |

**Example:**
``` go
config.GetFloatSlice("floatSlice") // [1.42 3.24]
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

You can set a config value at runtime with the `config.Set(key, value)` function. Bear in mind that this change is **temporary** and will be lost after your application restarts or if the config is reloaded. This function is mainly used for testing purposes. Values set using this function are still being validated: your application will panic and revert changes if the validation doesn't pass.

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

<p><Badge text="Since v3.0.0"/></p>

Configuration can be expanded. It is very likely that a plugin or a package you're developing is using some form of options. These options can be added to the configuration system so it is not needed to set them in the code or to make some wiring.

#### config.Register

Register a new config entry and its validation.

Each module should register its config entries in an `init()` function, even if they don't have a default value, in order to ensure they will be validated.

Each module should use its own category and use a name both expressive and unique to avoid collisions. For example, the `auth` package registers, among others, `auth.basic.username` and `auth.jwt.expiry`, thus creating a category for its package, and two subcategories for its features.

To register an entry without a default value (only specify how it will be validated), set `Entry.Value` to `nil`.

Is `IsSlice` is `true`, the value of the entry will be a slice of the given `Type`. Authorized values for slices define the values that can be used inside the slice. It doesn't represent the value of the slice itself (content and order).

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
    IsSlice:          false,
    AuthorizedValues: []interface{}{},
  })
  
  // Without a default value (only validation)
  config.Register("my-plugin.protocol", config.Entry{
    Value:            nil,
    Type:             reflect.String,
    IsSlice:          false,
    AuthorizedValues: []interface{}{"ftp", "sftp", "scp"},
  })

  // Slice
  config.Register("my-plugin.remoteHosts", config.Entry{
    Value:            []string{"first host", "second host"},
    Type:             reflect.String,
    IsSlice:          true,
    AuthorizedValues: []interface{}{},
  })
}
```

## Configuration reference

### App category

| Entry           | Type     | Default     | Note                                                                                                           |
|-----------------|----------|-------------|----------------------------------------------------------------------------------------------------------------|
| name            | `string` | "goyave"    |                                                                                                                |
| environment     | `string` | "localhost" |                                                                                                                |
| debug           | `bool`   | `true`      | When activated, print stacktrace on error and sends error message in response. **Disable this in production!** |
| defaultLanguage | `string` | "en-US"     | See the [Localization](./advanced/localization.html) section                                                   |

### Server category

| Entry         | Type      | Accepted values | Default     | Note                                                                      |
|---------------|-----------|-----------------|-------------|---------------------------------------------------------------------------|
| host          | `string`  | any             | "127.0.0.1" |                                                                           |
| domain        | `string`  | any             | ""          | Used for URL generation. Leave empty to use IP instead.                   |
| protocol      | `string`  | "http", "https" | "http"      | See the [HTTPS](#setting-up-https) section                                |
| port          | `int`     | any             | `8080`      |                                                                           |
| httpsPort     | `int`     | any             | `8081`      |                                                                           |
| timeout       | `int`     | any             | `10`        | Timeout in seconds                                                        |
| maxUploadSize | `float64` | any             | `10`        | Maximum size of the request, in MiB                                       |
| maintenance   | `bool`    | any             | `false`     | If `true`, start the server in maintenance mode. (Always return HTTP 503) |

#### TLS sub-category

| Entry | Type     | Default | Note                  |
|-------|----------|---------|-----------------------|
| cert  | `string` | none    | Path to your TLS cert |
| key   | `string` | none    | Path to your TLS key  |

### Database category

| Entry              | Type     | Default                                                                 | Note                                                      |
|--------------------|----------|-------------------------------------------------------------------------|-----------------------------------------------------------|
| connection         | `string` | "none"                                                                  | See the [Database](./basics/database.html) guide          |
| host               | `string` | "127.0.0.1"                                                             |                                                           |
| port               | `int`    | `3306`                                                                  |                                                           |
| name               | `string` | "goyave"                                                                |                                                           |
| username           | `string` | "root"                                                                  |                                                           |
| password           | `string` | "root"                                                                  |                                                           |
| otions             | `string` | "charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true&loc=Local" |                                                           |
| maxOpenConnections | `int`    | `20`                                                                    |                                                           |
| maxIdleConnections | `int`    | `20`                                                                    |                                                           |
| maxLifetime        | `int`    | `300`                                                                   | The maximum time (in seconds) a connection may be reused. |
| autoMigrate        | `bool`   | `false`                                                                 | When activated, migrate all registered models at startup  |


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
