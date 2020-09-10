---
meta:
  - name: "og:title"
    content: "Localization - Goyave"
  - name: "twitter:title"
    content: "Localization - Goyave"
  - name: "title"
    content: "Localization - Goyave"
---

# Localization

[[toc]]

## Introduction

The Goyave framework provides a convenient way to support multiple languages within your application. Out of the box, Goyave only provides the `en-US` language.

## Writing language files

Language files are stored in the `resources/lang` directory.

:::vue
.
└── resources
    └── lang
        └── en-US (*language name*)
            ├── fields.json (*optional*)
            ├── locale.json (*optional*)
            └── rules.json (*optional*)
:::

Each language has its own directory and should be named with an [ISO 639-1](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) language code. You can also append a variant to your languages: `en-US`, `en-UK`, `fr-FR`, `fr-CA`, ... **Case is important.**

Each language directory contains three files. Each file is **optional**.
- `fields.json`: field names translations and field-specific rule messages.
- `locale.json`: all other language lines.
- `rules.json`: validation rules messages.

::: tip
All directories in the `resources/lang` directory are automatically loaded when the server starts.
:::

### Fields

The `fields.json` file contains the field names translations and their rule-specific messages. Translating field names helps making more expressive messages instead of showing the technical field name to the user. Rule-specific messages let you override a validation rule message for a specific field.

**Example:**
``` json
{
    "email": {
        "name": "email address",
        "rules": {
            "required": "You must provide an :field."
        }
    }
}
```

This `fields.json` file will change the validation message of the `required` validation rule to `You must provide an email address`.

::: tip
Learn more about validation messages placeholders in the [validation](../basics/validation.html) section.
:::

### Locale

The `locale.json` file contains all language lines that are not related to validation. This is the place where you should write the language lines for your user interface or for the messages returned by your controllers.

**Example:**
``` json
{
    "product.created": "The product have been created with success.",
    "product.deleted": "The product have been deleted with success."
}
```
::: tip
It is a good practice to use **dot-separated** names for language lines to help making them clearer and more expressive.
:::

### Rules

The `rules.json` file contains the validation rules messages. These messages can have **[placeholders](../basics/validation.html#placeholders)**, which will be automatically replaced by the validator with dynamic values. If you write custom validation rules, their messages shall be written in this file.

**Example:**

``` json
{
    "integer": "The :field must be an integer.",
    "starts_with": "The :field must start with one of the following values: :values.",
    "same": "The :field and the :other must match."
}
```

#### Type-dependent rules

The following rules have **type-dependent** messages. That means that their message is different depending on the type of the validated data.

- `min`
- `max`
- `size`
- `greater_than`
- `greater_than_equal`
- `lower_than`
- `lower_than_equal`
- `between`

Type-dependent rules must have a language line for the four following types:
- `string`
- `numeric`
- `array`
- `file` 

**Example:**
```json
{
    "min.string": "The :field must be at least :min characters.",
    "min.numeric": "The :field must be at least :min.",
    "min.array": "The :field must have at least :min items.",
    "min.file": "The :field must be at least :min KiB."
}
```

#### Array validation

Each rule, except the file-related rules and the `confirmed` rule, can be used to validate array values. If a rule is used to validate an array value and doesn't pass, the rule message `validation.rules.<rule_name>.array` (or `validation.rules.<rule_name>.<type>.array` if the rule is type-dependent) is returned.

**Example:**
```json
{
    "min.string.array": "The :field values must be at least :min characters.",
    "min.numeric.array": "The :field values must be at least :min.",
    "min.array.array": "The :field values must have at least :min items.",
    "digits.array": "The :field values must be digits only."
}
```

### Overrides

If you define the `en-US`  language in your application, the default language lines will be overridden by the ones in your language files, and all the undefined ones will be kept.

It is possible to load a language directory manually from another location than the stardard `resources/lang` using the `lang.Load()` function. If the loaded language is already available in your application, the newly loaded one will override the previous in the same manner.

## Using localization

When an incoming request enters your application, the core language middleware checks if the `Accept-Language` header is set, and set the `goyave.Request`'s `Lang` attribute accordingly. Localization is handled automatically by the validator.

To use the localization feature, import the `lang` package:
``` go
import "github.com/System-Glitch/goyave/v3/lang"
```

The main function of the localization feature is `lang.Get(language, line string)`. This function lets you retrieve a language entry.

For validation rules and attributes messages, use the following dot-separated paths:
- `validation.rules.<rule_name>`
- `validation.rules.<rule_name>.string`
- `validation.rules.<rule_name>.numeric`
- `validation.rules.<rule_name>.array`
- `validation.rules.<rule_name>.file`
- `validation.fields.<field_name>`
- `validation.fields.<field_name>.<rule_name>`

For normal lines, just use the name of the line. Note that if you have a line called "validation", it won't conflict with the dot-separated paths. If the line cannot be found, or the requested language is not available, the function will return the exact `line` attribute.

**Example:**
``` go
func ControllerHandler(response *goyave.Response, request *goyave.Request) {
    response.String(http.StatusOK, lang.Get(request.Lang, "my-custom-message"))
}
```

### Placeholders

<p><Badge text="Since v2.10.0"/></p>

Language lines can contain **placeholders**. Placeholders are identified by a colon directly followed by the placeholder name:

```json
"greetings": "Greetings, :username!"
```

The last parameter of the `lang.Get()` method is a variadic associative slice of placeholders and their replacement. In the following example, the placeholder `:username` will be replaced with the Name field in the user struct.

```go
lang.Get("en-US", "greetings", ":username", user.Name) // "Greetings, Taylor!"
```

You can provide as many as you want:

```go
lang.Get("en-US", "greetings-with-date", ":username", user.Name, ":day", "Monday") // "Greetings, Taylor! Today is Monday"
```

::: tip
When a placeholder is given, **all occurrences** are replaced.

```json
"popular": ":product are very popular. :product sales exceeded 1000 last week."
```
```go
lang.Get("en-US", "popular", ":product", "Lawnmowers")
// "Lawnmowers are very popular. Lawnmowers sales exceeded 1000 last week."
```
:::

### Localization reference

::: table
[Get](#lang-get)
[Load](#lang-load)
[IsAvailable](#lang-isavailable)
[GetAvailableLanguages](#lang-getavailablelanguages)
[DetectLanguage](#lang-detectlanguage)
:::

#### lang.Get

Get a language line.

| Parameters               | Return   |
|--------------------------|----------|
| `lang string`            | `string` |
| `line string`            |          |
| `placeholders ...string` |          |

**Example:**
``` go
fmt.Println(lang.Get("en-US", "my-custom-message")) // "my message"
fmt.Println(lang.Get("en-US", "validation.rules.greater_than.string")) // "The :field must be longer than the :other."
fmt.Println(lang.Get("en-US", "validation.fields.email")) // "email address"
fmt.Println(lang.Get("en-US", "greetings", ":username", user.Name)) // "Greetings, Taylor!"
```

#### lang.Load

Load a language directory.

| Parameters        | Return |
|-------------------|--------|
| `language string` | `void` |
| `path string`     |        |

**Example:**
``` go
lang.Load("zh", "/path/to/chinese-lang")
```

#### lang.IsAvailable

Returns true if the language is available.

| Parameters    | Return |
|---------------|--------|
| `lang string` | `bool` |

**Example:**
``` go
fmt.Println(lang.IsAvailable("zh")) // true
```

#### lang.GetAvailableLanguages

Returns a slice of all loaded languages.

This can be used to generate different routes for all languages supported by your applications such as:
```
/en/products
/fr/produits
...
```

| Parameters    | Return |
|---------------|--------|
| `lang string` | `bool` |

**Example:**
``` go
fmt.Println(lang.GetAvailableLanguages()) // [en-US zh]
```

#### lang.DetectLanguage

DetectLanguage detects the language to use based on the given lang string.
The given lang string can use the HTTP "Accept-Language" header format.

- If `*` is provided, the default language will be used.
- If multiple languages are given, the first available language will be used, and if none are available, the default language will be used.
- If no variant is given (for example "en"), the first available variant will be used.
 
For example, if `en-US` and `en-UK` are available and the request accepts `en`, `en-US` will be used.

| Parameters    | Return   |
|---------------|----------|
| `lang string` | `string` |

**Example:**
``` go
fmt.Println(lang.DetectLanguage("en, fr-FR;q=0.9")) // "en-US"
```