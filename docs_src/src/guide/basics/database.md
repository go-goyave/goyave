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
  "gorm.io/gorm"
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
import _ "github.com/System-Glitch/goyave/v3/database/dialect/mysql"
// import _ "github.com/System-Glitch/goyave/v3/database/dialect/postgres"
// import _ "github.com/System-Glitch/goyave/v3/database/dialect/sqlite"
// import _ "github.com/System-Glitch/goyave/v3/database/dialect/mssql"
```

::: tip
For SQLite, only the `database.name` config entry is required.
:::

---

You can **register more dialects** for GORM. Start by implementing or importing it, then tell Goyave how to build the connection string for this dialect:

```go
import (
  "github.com/System-Glitch/goyave/v3/database"
  "example.com/user/mydriver"
)

func init() {
  database.RegisterDialect("my-driver", "{username}:{password}@({host}:{port})/{name}?{options}", mydriver.Open)
}
```

::: tip
See the [GORM "Write driver" documentation](https://gorm.io/docs/write_driver.html).
:::

Template format accepts the following placeholders, which will be replaced with the corresponding configuration entries automatically:
- `{username}`
- `{password}`
- `{host}`
- `{port}`
- `{name}`
- `{options}`

You cannot override a dialect that already exists.

#### database.RegisterDialect

| Parameters                         | Return |
|------------------------------------|--------|
| `name string`                      | `void` |
| `template string`                  |        |
| `initializer DialectorInitializer` |        |

::: tip
`DialectorInitializer` is an alias for `func(dsn string) gorm.Dialector`
:::

## Getting a database connection

#### database.GetConnection

Returns the global database connection pool. Creates a new connection pool if no connection is available.

By default, the [`PrepareStmt`](https://gorm.io/docs/performance.html#Caches-Prepared-Statement) option is **enabled**.

The connections will be closed automatically on server shutdown so you don't need to call `Close()` when you're done with the database.


| Parameters | Return     |
|------------|------------|
|            | `*gorm.DB` |

**Example:**
``` go
db := database.GetConnection()
db.First(&user)
```

#### database.Conn

`Conn()` is a short alias for `GetConnection()`.

| Parameters | Return     |
|------------|------------|
|            | `*gorm.DB` |

**Example:**
``` go
db := database.Conn()
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

You can modify the global instance of `*gorm.DB` when it's created (and re-created, after a `Close()` for example) using `Initializer` functions. This is useful if you want to set global settings such as `gorm:table_options` and make them effective for you whole application. It is recommended to register initializers **before** starting the application.

Initializer functions are called in order, meaning that functions added last can override settings defined by previous ones.

```go
database.AddInitializer(func(db *gorm.DB) {
    db.Config.SkipDefaultTransaction = true
    db.Statement.Settings.Store("gorm:table_options", "ENGINE=InnoDB")
})
```

#### database.AddInitializer

| Parameters                         | Return |
|------------------------------------|--------|
| `initializer database.Initializer` | `void` |

::: tip
- `database.Initializer` is an alias for `func(*gorm.DB)`
- Useful link related to initializers: [GORM config](https://gorm.io/docs/gorm_config.html)
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
    Email        string  `gorm:"type:varchar(100);uniqueIndex"`
    Role         string  `gorm:"size:255"` // set field size to 255
    MemberNumber *string `gorm:"unique;not null"` // set member number to unique and not null
    Num          int     `gorm:"autoIncrement"` // set num to auto incrementable
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

Sometimes you may wish to exclude some fields from your model's JSON form, such as passwords. To do so, you can add the `json:"-"` tag to the field you want to hide.

``` go
type User struct {
    Username string
    Password string `json:"-"`
}
```

The problem with `json:"-"` is that you won't be able to see those fields if you decide to serialize your records in json for another use, such as statistics. In that case, you can add the `model:"hide"` tag to the field you want to hide, and use [`helper.RemoveHiddenFields()`](../advanced/helpers.html#helper-removehiddenfields) to filter out those fields.

``` go
type User struct {
    Username string
    Password string `model:"hide" json:",omitempty"`
}
```

### Automatic migrations

If the `database.autoMigrate` config option is set to true, all registered models will be automatically migrated when the server starts.

::: warning
Automatic migrations create tables, missing foreign keys, constraints, columns and indexes, and will change existing column’s type if it’s size, precision or nullable changed. They **wont't** delete unused columns.
:::

If you would like to know more about migrations using Gorm, read their [documentation](https://gorm.io/docs/migration.html).

## Pagination

<p><Badge text="Since v3.4.0"/></p>

`database.Paginator` is a tool that helps you paginate records. This structure contains pagination information (current page, maximum page, total number of records), which is automatically fetched. You can send the paginator directly to the client as a response.

**Example:**
```go
articles := []model.Article{}
db := database.Conn()
paginator := database.NewPaginator(db, page, pageSize, &articles)
result := paginator.Find()
if response.HandleDatabaseError(result) {
    response.JSON(http.StatusOK, paginator)
}
```

When calling `paginator.Find()`, the `paginator` struct will be automatically updated with the total and max pages. The destination slice passed to `NewPaginator()` is also updated automatically. (`articles` in the above example)

---

You can add clauses to your SQL query before creating the paginator. This is especially useful if you want to paginate search results. The condition will be applied to both the total records count query and the actual page query.

**Full example:**
```go
func Index(response *goyave.Response, request *goyave.Request) {
    articles := []model.Article{}
    page := 1
    if request.Has("page") {
        page = request.Integer("page")
    }
    pageSize := DefaultPageSize
    if request.Has("pageSize") {
        pageSize = request.Integer("pageSize")
    }

    tx := database.Conn()

    if request.Has("search") {
        search := helper.EscapeLike(request.String("search"))
        tx = tx.Where("title LIKE ?", "%"+search+"%")
    }

    paginator := database.NewPaginator(tx, page, pageSize, &articles)
    result := paginator.Find()
    if response.HandleDatabaseError(result) {
        response.JSON(http.StatusOK, paginator)
    }
}
```

#### database.NewPaginator

Create a new `Paginator`.

Given DB transaction can contain clauses already, such as WHERE, if you want to filter results.

| Parameters         | Return       |
|--------------------|--------------|
| `db *gorm.DB`      | `*Paginator` |
| `page int`         |              |
| `pageSize int`     |              |
| `dest interface{}` |              |

**`Paginator` definition:**
```go
type Paginator struct {
	  MaxPage     int64
	  Total       int64
	  PageSize    int
	  CurrentPage int
	  Records     interface{}
}
```

#### paginator.Find

Find requests page information (total records and max page) and executes the transaction. The Paginate struct is updated automatically, as well as the destination slice given in `NewPaginate()`.

| Parameters | Return     |
|------------|------------|
|            | `*gorm.DB` |

## Setting up SSL/TLS

### MySQL

If you want to make your database connection use a TLS configuration, create `database/tls.go`. In this file, create an `init()` function which will load your certificates and keys.

Don't forget to blank import the database package in your `kernel.go`: `import _ "myproject/database"`. Finally, for a configuration named "custom", add `&tls=custom` at the end of the `database.options` configuration entry.

```go
package database

import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"

    "github.com/System-Glitch/goyave/v3"
    "github.com/go-sql-driver/mysql"
)

func init() {
    rootCertPool := x509.NewCertPool()
    pem, err := ioutil.ReadFile("/path/ca-cert.pem")
    if err != nil {
        goyave.ErrLogger.Fatal(err)
    }
    if ok := rootCertPool.AppendCertsFromPEM(pem); !ok {
        goyave.ErrLogger.Fatal("Failed to append PEM.")
    }
    clientCert := make([]tls.Certificate, 0, 1)
    certs, err := tls.LoadX509KeyPair("/path/client-cert.pem", "/path/client-key.pem")
    if err != nil {
        goyave.ErrLogger.Fatal(err)
    }
    clientCert = append(clientCert, certs)
    mysql.RegisterTLSConfig("custom", &tls.Config{
        RootCAs:      rootCertPool,
        Certificates: clientCert,
    })
}
```
[Reference](https://pkg.go.dev/github.com/go-sql-driver/mysql#RegisterTLSConfig)

### PostgreSQL

For PostgreSQL, you only need to add a few options to the `database.options` configuration entry.

```
sslmode=verify-full sslrootcert=root.crt sslkey=client.key sslcert=client.crt
```

Replace `root.crt`, `client.key` and `client.crt` with the paths to the corresponding files.

[Reference](https://pkg.go.dev/github.com/lib/pq#hdr-Connection_String_Parameters)

### MSSQL

Refer to the [driver's documentation](https://github.com/denisenkom/go-mssqldb#less-common-parameters).
