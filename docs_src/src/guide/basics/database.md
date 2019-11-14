# Database

[[toc]]

## Introduction

Most web applications use a database. In this section, we are going to see how Goyave applications can query a database, using the awesome [Gorm ORM](https://gorm.io/).

Database connections are managed by the framework and are long-lived. When the server shuts down, the database connections are closed automatically. So you don't have to worry about creating, closing or refreshing database connections in your application.

All functions below require the `database` package to be imported.

``` go
import "github.com/System-Glitch/goyave/database"
```

## Configuration

Very few code is required to get started with databases. There are some [configuration](../configuration#configuration-reference) options that you need to change though:
- `dbConnection`
- `dbHost`
- `dbPort`
- `dbName`
- `dbUsername`
- `dbPassword`
- `dbOptions`
- `dbMaxOpenConnection`
- `dbMaxIdleConnection`

::: tip
`dbOptions` represents the addtional connection options. For example, when using MySQL, you should use the `parseTime=true` option so `time.Time` can be handled correctly. Available options differ from one driver to another and can be found in their respective documentation.
:::

### Drivers

The framework supports the following sql drivers:
- `none` (*Disable database features*)
- `mysql`
- `postgres`
- `sqlite3`
- `mssql`

Change the `dbConnection` config entry to the desired driver.

In order to be able connect to the database, Gorm needs a database driver to be imported. Add the following import to your `kernel.go`:
``` go
import _ "github.com/jinzhu/gorm/dialects/mysql"
// import _ "github.com/jinzhu/gorm/dialects/postgres"
// import _ "github.com/jinzhu/gorm/dialects/sqlite"
// import _ "github.com/jinzhu/gorm/dialects/mssql"
```

::: tip
For SQLite, only the `dbName` config entry is required.
:::

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

| Parameters | Return |
|------------|--------|
|            | `void` |

**Example:**
``` go
database.Close()
```

## Models

A model is a structure reflecting a database table structure. An instance of a model is a single database record. Each model is defined in its own file inside the `database/models` directory.

### Defining a model

Models are usually just normal Golang structs, basic Go types, or pointers of them. `sql.Scanner` and `driver.Valuer` interfaces are also supported.

```go
func init() {
    goyave.RegisterModel(&User{})
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
import _ "database/models"
```
:::

Learn more about model declaration in the [Gorm documentation](https://gorm.io/docs/models.html).

### Automatic migrations

If the `dbAutoMigrate` config option is set to true, all registered models will be automatically migrated when the server starts.

::: warning
Automatic migrations **only creates** tables. Missing columns and indexes won't be created, modified columns won't be changed and unused columns won't be deleted.
:::

If you would like to know more about migrations using Gorm, read their [documentation](https://gorm.io/docs/migration.html).
