---
meta:
  - name: "og:title"
    content: "Helpers - Goyave"
  - name: "twitter:title"
    content: "Helpers - Goyave"
  - name: "title"
    content: "Helpers - Goyave"
---

# Helpers

[[toc]]

The Goyave framework offers a collection of helpers to ease development.

## General

The helpers require the `helper` package to be imported.

``` go
import "github.com/System-Glitch/goyave/v3/helper"
```

**List of general helpers**:
::: table
[IndexOf](#helper-indexof)
[Contains](#helper-contains)
[IndexOfStr](#helper-indexofstr)
[ContainsStr](#helper-containsstr)
[SliceEqual](#helper-sliceequal)
[ToFloat64](#helper-tofloat64)
[ToString](#helper-tostring)
[ParseMultiValuesHeader](#helper-parsemultivaluesheader)
[RemoveHiddenFields](#helper-removehiddenfields)
[Only](#helper-only)
[EscapeLike](#helper-escapelike)
:::

#### helper.IndexOf

Get the index of the given value in the given slice, or `-1` if not found.

| Parameters          | Return |
|---------------------|--------|
| `slice interface{}` | `int`  |
| `value interface{}` |        |

**Example:**
``` go
slice := []interface{}{'r', "Goyave", 3, 2.42}
fmt.Println(helper.IndexOf(slice, "Goyave")) // 1
```

#### helper.Contains

Check if a generic slice contains the given value.

| Parameters          | Return |
|---------------------|--------|
| `slice interface{}` | `bool` |
| `value interface{}` |        |

**Example:**
``` go
slice := []interface{}{'r', "Goyave", 3, 2.42}
fmt.Println(helper.Contains(slice, "Goyave")) // true
```

#### helper.IndexOfStr

Get the index of the given value in the given string slice, or `-1` if not found.

Prefer using this helper instead of `IndexOf` for better performance.

| Parameters       | Return |
|------------------|--------|
| `slice []string` | `int`  |
| `value []string` |        |

**Example:**
``` go
slice := []string{"Avogado", "Goyave", "Pear", "Apple"}
fmt.Println(helper.IndexOfStr(slice, "Goyave")) // 1
```

#### helper.ContainsStr

Check if a string slice contains the given value.

Prefer using this helper instead of `Contains` for better performance.

| Parameters       | Return |
|------------------|--------|
| `slice []string` | `bool` |
| `value []string` |        |

**Example:**
``` go
slice := []string{"Avogado", "Goyave", "Pear", "Apple"}
fmt.Println(helper.ContainsStr(slice, "Goyave")) // true
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

Convert a numeric value to `float64`.

| Parameters          | Return    |
|---------------------|-----------|
| `value interface{}` | `float64` |
|                     | `error`   |

**Examples:**
``` go
fmt.Println(helper.ToFloat64(1.42)) // 1.42 nil
fmt.Println(helper.ToFloat64(1)) // 1.0 nil
fmt.Println(helper.ToFloat64("1.42")) // 1.42 nil
fmt.Println(helper.ToFloat64("NaN")) // NaN nil
fmt.Println(helper.ToFloat64([]string{})) // 0 'strconv.ParseFloat: parsing "[]": invalid syntax'
```

#### helper.ToString

Convert a generic value to string.

| Parameters          | Return   |
|---------------------|----------|
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

#### helper.RemoveHiddenFields

Remove hidden fields if the given model is a struct pointer. All fields marked with the tag `model:"hide"` will be set to their zero value.

| Parameters          | Return |
|---------------------|--------|
| `model interface{}` | `void` |

**Example:**
``` go
type Model struct {
    Username string
    Password string `model:"hide" json:",omitempty"`
}

model := &Model{
    Username: "Jeff",
    Password: "bcrypted password",
}

helper.RemoveHiddenFields(model)
fmt.Println(model) // &{ Jeff}
```

#### helper.Only

Extracts the requested field from the given `map[string]` or structure and returns a `map[string]interface{}` containing only those values.

| Parameters         | Return                   |
|--------------------|--------------------------|
| `data interface{}` | `map[string]interface{}` |
| `fields ...string` |                          |

**Example:**
``` go
type Model struct {
  Field string
  Num   int
  Slice []float64
}
model := Model{
  Field: "value",
  Num:   42,
  Slice: []float64{3, 6, 9},
}
res := Only(model, "Field", "Slice")
```

Result:
```go
map[string]interface{}{
  "Field": "value",
  "Slice": []float64{3, 6, 9},
}
```

#### helper.EscapeLike

Escape "%" and "_" characters in the given string for use in SQL "LIKE" clauses.

| Parameters   | Return   |
|--------------|----------|
| `str string` | `string` |

**Example:**
``` go
search := helper.EscapeLike("se%r_h")
fmt.Println(search) // "se\%r\_h"
```

## Filesystem

The filesystem helpers require the `helper/filesystem`  package to be imported.

``` go
import "github.com/System-Glitch/goyave/v3/helper/filesystem"
```

All files received in a request are stored in the `filesystem.File` structure. This structres gives all the information you need on a file and its content, as well as a helper function to save it easily.

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

Creates directories if needed.

Returns the actual file name.

| Parameters    | Return   |
|---------------|----------|
| `path string` | `string` |
| `name string` |          |

**Example:**
``` go
image := request.File("image")[0]
// As file fields can be multi-files uploads, a file field
// is always a slice.

name := request.String("name")
product := model.Product{
    Name: name,
    Price: request.Numeric("price"),
    Image: image.Save("storage/img", name)
}
database.Conn().Create(&product)
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
fmt.Println(filesystem.GetFileExtension("config.json")) // "application/json; charset=utf-8"
fmt.Println(filesystem.GetFileExtension("index.html")) // "text/html; charset=utf-8"
```

#### filesystem.FileExists

Returns true if the file at the given path exists and is readable. Returns false if the given file is a directory.

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
