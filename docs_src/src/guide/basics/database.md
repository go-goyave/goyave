---
meta:
  - name: "og:title"
    content: "Database - Goyave"
  - name: "twitter:title"
    content: "Database - Goyave"
  - name: "title"
    content: "Database - Goyave"
---

# Database

[[toc]]

## Introduction

Most web applications use a database. In this section, we are going to see how Goyave applications can query a database, using the awesome [Gorm ORM](https://gorm.io/).

Database connections are managed by the framework and are long-lived. When the server shuts down, the database connections are closed automatically. So you don't have to worry about creating, closing or refreshing database connections in your application.

All functions below require the `database` and the `gorm` packages to be imported.

``` go
import (
  "github.com/System-Glitch/goyave/v3/database"
  "github.com/jinzhu/gorm"
)
```

## Configuration

Very few code is required to get started with databases. There are some [configuration](../configuration.html#database-category) options that you need to change though:
- `database.connection`
- `database.host`
- `database.port`
- `database.name`
- `database.username`
- `database.password`
- `database.options`
- `database.maxOpenConnection`
- `database.maxIdleConnection`
- `database.maxLifetime`

::: tip
`database.options` represents the additional connection options. For example, when using MySQL, you should use the `parseTime=true` option so `time.Time` can be handled correctly. Available options differ from one driver to another and can be found in their respective documentation.
:::

### Drivers

The framework supports the following sql drivers out-of-the-box:
- `none` (*Disable database features*)
- `mysql`
- `postgres`
- `sqlite3`
- `mssql`

Change the `database.connection` config entry to the desired driver.

In order to be able connect to the database, Gorm needs a database driver to be imported. Add the following import to your `kernel.go`:
``` go
import _ "github.com/jinzhu/gorm/dialects/mysql"
// import _ "github.com/jinzhu/gorm/dialects/postgres"
// import _ "github.com/jinzhu/gorm/dialects/sqlite"
// import _ "github.com/jinzhu/gorm/dialects/mssql"
```

::: tip
For SQLite, only the `database.name` config entry is required.
:::

---

You can **register more dialects** for GORM [like you would usually do](http://gorm.io/docs/dialects.html). There is one more step required: you need to tell Goyave how to build the connection string for this dialect:

```go
import (
  "github.com/System-Glitch/goyave/v3/database"
  "github.com/jinzhu/gorm"
  _ "example.com/user/my-dialect"
)

type myDialect struct{
  db gorm.SQLCommon
  gorm.DefaultForeignKeyNamer
}

// Dialect implementation...

func init() {
  gorm.RegisterDialect("my-dialect", &myDialect{})
  database.RegisterDialect("my-dialect", "{username}:{password}@({host}:{port})/{name}?{options}")
}
```


Template format accepts the following placeholders, which will be replaced with the corresponding configuration entries automatically:
- `{username}`
- `{password}`
- `{host}`
- `{port}`
- `{name}`
- `{options}`

You cannot override a dialect that already exists.

#### database.RegisterDialect

| Parameters        | Return |
|-------------------|--------|
| `name string`     | `void` |
| `template string` |        |


## Getting a database connection

#### database.GetConnection

Returns the global database connection pool. Creates a new connection pool if no connection is available.

The connections will be closed automatically on server shutdown so you don't need to call `Close()` when you're done with the database.


| Parameters | Return     |
|------------|------------|
|            | `*gorm.DB` |

**Example:**
``` go
db := database.GetConnection()
db.First(&user)
```

::: tip
Learn how to use the CRUD interface and the query builder in the [Gorm documentation](https://gorm.io/docs/index.html).
:::

#### database.Close

If you want to manually close the database connection, you can do it using `Close()`. New connections can be re-opened using `GetConnection()` as usual. This function does nothing if the database connection is already closed or has never been created.

| Parameters | Return  |
|------------|---------|
|            | `error` |

**Example:**
``` go
database.Close()
```

### Connection initializers

You can modify the global instance of `*gorm.DB` when it's created (and re-created, after a `Close()` for example) using `Initializer` functions. This is useful if you want to set global settings such as `gorm:auto_preload` and make them effective for you whole application. It is recommended to register initializers **before** starting the application.

In your initializers, use `db.InstantSet()` and not `db.Set()`, since the latter clones the `gorm.DB` instance instead of modifying it.

Initializer functions are called in order, meaning that functions added last can override settings defined by previous ones.

```go
database.AddInitializer(func(db *gorm.DB) {
    db.InstantSet("gorm:auto_preload", true)
})
```

#### database.AddInitializer

| Parameters                         | Return |
|------------------------------------|--------|
| `initializer database.Initializer` | `void` |

::: tip
`database.Initializer` is an alias for `func(*gorm.DB)`
:::

#### database.ClearInitializers

Remove all database connection initializer functions.

| Parameters | Return |
|------------|--------|
|            | `void` |


## Models

A model is a structure reflecting a database table structure. An instance of a model is a single database record. Each model is defined in its own file inside the `database/model` directory.

### Defining a model

Models are usually just normal Golang structs, basic Go types, or pointers of them. `sql.Scanner` and `driver.Valuer` interfaces are also supported.

```go
func init() {
    database.RegisterModel(&User{})
}

type User struct {
    gorm.Model
    Name         string
    Age          sql.NullInt64
    Birthday     *time.Time
    Email        string  `gorm:"type:varchar(100);unique_index"`
    Role         string  `gorm:"size:255"` // set field size to 255
    MemberNumber *string `gorm:"unique;not null"` // set member number to unique and not null
    Num          int     `gorm:"AUTO_INCREMENT"` // set num to auto incrementable
    Address      string  `gorm:"index:addr"` // create index with name `addr` for address
    IgnoreMe     int     `gorm:"-"` // ignore this field
}
```

::: tip
All models should be **registered** in an `init()` function inside their model file. To ensure the `init()` functions are executed before the server starts, import the `models` package in your `kernel.go`.

``` go
import _ "database/model"
```
:::

Learn more about model declaration in the [Gorm documentation](https://gorm.io/docs/models.html).

#### database.RegisterModel

Registers a model for auto-migration.

| Parameters          | Return |
|---------------------|--------|
| `model interface{}` | `void` |

#### database.GetRegisteredModels

Get the registered models. The returned slice is a copy of the original, so it cannot be modified.

| Parameters | Return          |
|------------|-----------------|
|            | `[]interface{}` |

#### database.ClearRegisteredModels

Unregister all models.

| Parameters | Return |
|------------|--------|
|            | `void` |

### Hidden fields

<p><Badge text="Since v2.9.0"/></p>

Sometimes you may wish to exclude some fields from your model's JSON form, such as passwords. To do so, you can add the `model:"hide"` tag to the field you want to hide.

``` go
type User struct {
    Username string
    Password string `model:"hide" json:",omitempty"`
}
```

When a struct is sent as a response through `response.JSON()`, all its fields (including promoted fields) tagged with `model:"hide"` will be set to their zero value. Add the `json:",omitempty"` tag to entirely remove the field from the resulting JSON string.

You can also filter hidden fields by passing a struct to [`helper.RemoveHiddenFields()`](../advanced/helpers.html#helper-removehiddenfields).

### Automatic migrations

If the `database.autoMigrate` config option is set to true, all registered models will be automatically migrated when the server starts.

::: warning
Automatic migrations **only create** tables, missing columns and missing indexes. They **wont't change** existing column’s type or delete unused columns.
:::

If you would like to know more about migrations using Gorm, read their [documentation](https://gorm.io/docs/migration.html).
