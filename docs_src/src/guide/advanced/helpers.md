# Helpers

[[toc]]

The Goyave framework offers a collection of helpers to ease development.

## General

The helpers require the `helper` package to be imported.

``` go
import "github.com/System-Glitch/goyave/helper"
```

**List of general helpers**:
::: table
[Contains](#helper-contains)
[SliceEqual](#helper-sliceequal)
[ToFloat64](#helper-tofloat64)
[ToString](#helper-tostring)
[ParseMultiValuesHeader](#helper-parsemultivaluesheader)
:::

#### helper.Contains

Check if a generic slice contains the given value.

| Parameters          | Return |
|---------------------|--------|
| `slice interface{}` | `bool` |
| `value interface{}` |        |

**Example:**
``` go
slice := []string{"Avogado", "Goyave", "Pear", "Apple"}
fmt.Println(helper.Contains(slice, "Goyave")) // true
```

#### helper.SliceEqual

Check if two generic slices are the same.

| Parameters           | Return |
|----------------------|--------|
| `first interface{}`  | `bool` |
| `second interface{}` |        |

**Example:**
``` go
first := []string{"Avogado", "Goyave", "Pear", "Apple"}
second := []string{"Goyave", "Avogado", "Pear", "Apple"}
fmt.Println(helper.SliceEqual(first, second)) // false
```

#### helper.ToFloat64

Check if two generic slices are the same.

| Parameters          | Return    |
|---------------------|-----------|
| `value interface{}` | `float64` |
|                     | `error`   |

**Examples:**
``` go
fmt.Println(helper.ToFloat64(1.42)) // 1.42 nil
fmt.Println(helper.ToFloat64(1)) // 1.0 nil
fmt.Println(helper.ToFloat64("1.42")) // 1.42 nil
fmt.Println(helper.ToFloat64("NaN")) // 0 nil
fmt.Println(helper.ToFloat64([]string{})) // 0 nil
```

#### helper.ToString

Convert a generic value to string.

| Parameters          | Return    |
|---------------------|-----------|
| `value interface{}` | `string` |

**Examples:**
``` go
fmt.Println(helper.ToString(1.42)) // "1.42"
fmt.Println(helper.ToString(nil)) // "nil"
fmt.Println(helper.ToString("hello")) // "hello"
fmt.Println(helper.ToString([]string{})) // "[]"
```

#### helper.ParseMultiValuesHeader

Parses multi-values HTTP headers, taking the quality values into account. The result is a slice of values sorted according to the order of priority.

See: [https://developer.mozilla.org/en-US/docs/Glossary/Quality_values](https://developer.mozilla.org/en-US/docs/Glossary/Quality_values)

| Parameters      | Return                     |
|-----------------|----------------------------|
| `header string` | `[]filesystem.HeaderValue` |

**HeaderValue struct:**

| Attribute  | Type      |
|------------|-----------|
| `Value`    | `string`  |
| `Priority` | `float64` |

**Examples:**
``` go
fmt.Println(helper.ParseMultiValuesHeader("text/html,text/*;q=0.5,*/*;q=0.7"))
// [{text/html 1} {*/* 0.7} {text/* 0.5}]

fmt.Println(helper.ParseMultiValuesHeader("text/html;q=0.8,text/*;q=0.8,*/*;q=0.8"))
// [{text/html 0.8} {text/* 0.8} {*/* 0.8}]
```

## Filesystem

The filesystem helpers require the `filesystem`  package to be imported.

``` go
import "github.com/System-Glitch/goyave/helper/filesystem"
```

All files received in a requests are stored in the `filesystem.File` structure. This structres gives all the information you need on a file and its content, as well as a helper function to save it easily.

| Attribute  | Type                    |
|------------|-------------------------|
| `Header`   | `*multipart.FileHeader` |
| `MIMEType` | `string`                |
| `Data`     | `multipart.File`        |

::: warning
The data in `file.Header` come from the client and **shouldn't be trusted**. The filename is always optional and must not be used blindly by the application: path information should be stripped, and conversion to the server file system rules should be done. You cannot rely on the size given in the header neither.
:::

**List of filesystem helpers**:
::: table
[File.Save](#filesystem-file-save)
[GetFileExtension](#filesystem-getfileextension)
[GetMIMEType](#filesystem-getmimetype)
[FileExists](#filesystem-fileexists)
[IsDirectory](#filesystem-isdirectory)
[Delete](#filesystem-delete)
:::

#### filesystem.File.Save

Writes the given file on the disk.
Appends a timestamp to the given file name to avoid duplicate file names.
The file is not readable anymore once saved as its FileReader has already been closed.

Returns the actual path to the saved file.

| Parameters             | Return   |
|------------------------|----------|
| `path string`          | `string` |
| `name string`          |          |

**Example:**
``` go
image := request.Data["image"].([]filesystem.File)[0]
// As file fields can be multi-files uploads, a file field
// is always a slice.

name := request.Data["name"].(string)
product := model.Product{
    Name: name,
    Price: request.Data["price"].(float64),
    Image: image.Save("storage/img", name)
}
database.GetConnection().Create(&product)
```

#### filesystem.GetFileExtension

Returns the last part of a file name. If the file doesn't have an extension, returns an empty string.

| Parameters    | Return   |
|---------------|----------|
| `file string` | `string` |

**Examples:**
``` go
fmt.Println(filesystem.GetFileExtension("README.md")) // "md"
fmt.Println(filesystem.GetFileExtension("LICENSE")) // empty string
fmt.Println(filesystem.GetFileExtension("archive.tar.gz")) // "gz"
```

#### filesystem.GetMIMEType

Get the MIME type and size of the given file. If the file cannot be opened, panics. You should check if the file exists, using `filesystem.FileExists()`, before calling this function.

| Parameters    | Return   |
|---------------|----------|
| `file string` | `string` |

**Examples:**
``` go
fmt.Println(filesystem.GetMIMEType("logo.png")) // "image/png"
fmt.Println(filesystem.GetFileExtension("LICENSE")) // "application/octet-stream"
fmt.Println(filesystem.GetFileExtension("index.html")) // "text/html; charset=utf-8"
```

#### filesystem.FileExists

Returns true if the file at the given path exists and is readable. Returns false if the given file is a directory

| Parameters    | Return |
|---------------|--------|
| `file string` | `bool` |

**Example:**
``` go
fmt.Println(filesystem.FileExists("README.md")) // true
```

#### filesystem.IsDirectory

Returns true if the file at the given path exists, is a directory and is readable.

| Parameters    | Return |
|---------------|--------|
| `path string` | `bool` |

**Example:**
``` go
fmt.Println(filesystem.IsDirectory("README.md")) // false
fmt.Println(filesystem.IsDirectory("resources")) // true
```

#### filesystem.Delete

Delete the file at the given path. Panics if the file cannot be deleted.

You should check if the file exists, using `filesystem.FileExists()`, before calling this function.

| Parameters    | Return |
|---------------|--------|
| `file string` | `void` |

**Example:**
``` go
filesystem.Delete("README.md")
```
