# Validation

[[toc]]

## Introduction

Goyave provides a powerful, yet easy way to validate all incoming data, no matter its type or its format, thanks to a large number of validation rules.

Incoming requests are validated using **rules set**, which associate rules with each expected field in the request.

Validation rules can **alter the raw data**. That means that when you validate a field to be number, if the validation passes, you are ensured that the data you'll be using in your controller handler is a `float64`. Or if you're validating an IP, you get a `net.IP` object.

If a request contains a field with a `nil`/`null` value, and that this field doesn't have the `nullable` rule, the field is **removed** entirely from the request.

Validation is automatic. You just have to define a rules set and assign it to a route. When the validation doesn't pass, the request is stopped and the validation errors messages are sent as a response, using the correct [language](../advanced/localization). The HTTP response code of failed validation is **422 "Unprocessable Entity"**, or **400 "Bad Request"** if the body could not be parsed.

::: tip
You can customize the validation error messages by editing `resources/lang/<language>/rules.json`. Learn more in the [localization](../advanced/localization) section.
:::

The following features require the `validation` package to be imported.

``` go
import "github.com/System-Glitch/goyave/v2/validation"
```

## Rules sets

The `http/request` directory contains the requests validation rules sets. You should have one package per feature, regrouping all requests handled by the same controller. The package should be named `<feature_name>request`.

**Example:** (`http/request/productrequest/product.go`)
``` go
var (
	Store validation.RuleSet = validation.RuleSet{
		"name":  {"required", "string", "between:3,50"},
		"price": {"required", "numeric", "min:0.01"},
		"image": {"nullable", "file", "image", "max:2048", "count:1"},
    }
    
    // ...
)
```

::: warning
**The order in which you assign rules is important**, as rules are executed in this order. The rules checking for the type of the data should **always be first**, or after `required` and `nullable`.

If a field is not **required** and is missing from the request, **no rules are checked**!
:::

::: tip
`validation.RuleSet` is an alias for `map[string][]string`
:::

---

Once your rules sets are defined, you need to assign them to your routes. The rule set for a route is the last parameter of the route definition. Learn more about routing in the [dedicated section](./routing).

``` go
router.Route("POST", "/product", product.Store, productrequest.Store)
```

## Available validation rules

::: table
[Required](#required)
[Nullable](#nullable)
[Numeric](#numeric)
[Integer](#integer)
[Min](#min-value)
[Max](#max-value)
[Between](#between-min-max)
[Greater than](#greater-than-field)
[Greater than or equal](#greater-than-equal-field)
[Lower than](#lower-than-field)
[Lower than or equal](#lower-than-equal-field)
[String](#string)
[Array](#array-type)
[Distinct](#distinct)
[Digits](#digits)
[Regex](#regex-pattern)
[Email](#email)
[Size](#size-value)
[Alpha](#alpha)
[Alpha dash](#alpha-dash)
[Alpha numeric](#alpha-num)
[Starts with](#starts-with-value1)
[Ends with](#ends-with-value1)
[In](#in-value1-value2)
[Not in](#not-in-value1-value2)
[In array](#in-array-field)
[Not in array](#not-in-array-field)
[Timezone](#timezone)
[IP](#ip)
[IPv4](#ipv4)
[IPv6](#ipv6)
[JSON](#json)
[URL](#url)
[UUID](#uuid-version)
[Bool](#bool)
[Same](#same-field)
[Different](#different-field)
[Confirmed](#confirmed)
[File](#file)
[MIME](#mime-foo)
[Image](#image)
[Extension](#extension-foo)
[Count](#count-value)
[Count min](#count-min-value)
[Count max](#count-max-value)
[Count between](#count-between-min-max)
[Date](#date-format)
[Before](#before-date)
[Before or equal](#before-equal-date)
[After](#after-date)
[After or equal](#after-equal-date)
[Date equals](#date-equals-date)
[Date between](#date-between-date1-date2)
:::

#### required

The field under validation must be present.

If the field is a string, the string must not be empty. If a field is `null` and has the `nullable` rule, the `required` rules passes. As non-nullable fields are removed if they have a `null` value, the `required` rule doesn't pass if a field is `null` and doesn't have the `nullable` rule.

#### nullable

The field under validation can have a `nil`/`null` value. If this rule is missing from the rules set, the field will be **removed** if it is `null`. This rule is especially useful when working with JSON requests or with primitives that can contain null values.

Be sure to check if your field is not null before using it in your handlers.
``` go
// In this example, field is numeric
if val, exists := request.Data["field"]; exists && val != nil {
    field, _ := request.Data["field"].(float64)
}
```

#### numeric

The field under validation must be numeric. Strings that can be converted to numbers are accepted.
This rule converts the field to `float64` if it passes.

#### integer

The field under validation must be an integer. Strings that can be converted to an integer are accepted.
This rule converts the field to `int` if it passes.

#### min:value

Depending on its type, the field under validation must be at least `min`.
Strings, numerics, array, and files are evaluated using the same method as the [`size`](#size-value) rule.

#### max:value

Depending on its type, the field under validation must not be superior to `value`.
Strings, numerics, array, and files are evaluated using the same method as the [`size`](#size-value) rule.

#### between:min,max

Depending on its type, the field under validation must be between `min` and `max`.
Strings, numerics, array, and files are evaluated using the same method as the [`size`](#size-value) rule.

#### greater_than:field

The field under validation must be greater than the given `field`. The two fields must have the same type.
Strings, numerics, array, and files are evaluated using the same method as the [`size`](#size-value) rule.

#### greater_than_equal:field

The field under validation must be greater or equal to the given `field`. The two fields must have the same type.
Strings, numerics, array, and files are evaluated using the same method as the [`size`](#size-value) rule.

#### lower_than:field

The field under validation must be lower than the given `field`. The two fields must have the same type.
Strings, numerics, array, and files are evaluated using the same method as the [`size`](#size-value) rule.

#### lower_than_equal:field

The field under validation must be lower or equal to the given `field`. The two fields must have the same type.
Strings, numerics, array, and files are evaluated using the same method as the [`size`](#size-value) rule.

#### string

The field under validation must be a string.

#### array:type

The field under validation must be an array. The `type` parameter is **optional**.

If no type is provided, the field has the type `[]interface{}` after validation. If a type is provided, the array is converted to a slice of the correct type, and all values in the array are validated in the same way as standard fields.

For example, with the rule `array:url`, all values must be valid URLs and the field will be converted to `[]*url.URL`.

**Available types:**
- `string`
- `numeric`
- `integer`
- `timezone`
- `ip`, `ipv4`, `ipv6`
- `url`
- `uuid`
- `bool`
- `date`

::: tip
For the `uuid` and `date` types, you can pass a second parameter: `array:date,02-01-2006`
:::

#### distinct

The field under validation must be an array and have distinct values.

#### digits

The field under validation must be a string and contain only digits.

#### regex:pattern

The field under validation must be a string and match the given `pattern`.

#### email

The field under validation must be a string and be an email address.

::: warning
This rule is not enough to properly validate email addresses. The only way to ensure an email address is valid is by sending a confirmation email.
:::

#### size:value

Depending on its type, the field under validation must:
- Strings: have a length of `value` characters.
- Numerics: be equal to `value`.
- Arrays: exactly have `value` items.
- Files: weight exactly `value` KiB. 
    - *Note: for this rule only (not for `min`, `max`, etc), the size of the file under validation is **rounded** to the closest KiB.*
    - When the field is a multi-files upload, the size of **all files** is checked.

#### alpha

The field under validation must be a string and be entirely alphabetic characters.

#### alpha_dash

The field under validation must be a string and be entirely alphabetic-numeric characters, dashes or underscores.

#### alpha_num

The field under validation must be a string and be entirely alphabetic-numeric characters.

#### starts_with:value1,...

The field under validation must be a string and start with of the given values.

#### ends_with:value1,...

The field under validation must be a string and end with one of the given values.

#### in:value1,value2,...

The field under validation must be a one of the given values. Only numerics and strings are checked.

#### not_in:value1,value2,...

The field under validation must not be a one of the given values. Only numerics and strings are checked.

#### in_array:field

The field under validation must be a one of the values in the given `field`. Only numerics and strings are checked, and the given `field` must be an array.

#### not_in_array:field

The field under validation must not be a one of the values in the given `field`. Only numerics and strings are checked, and the given `field` must be an array.

#### timezone

The field under validation must be a string and be a valid timezone. This rule converts the field to `*time.Location` if it passes.

Valid timezones are:
- UTC
- Timezones from the IANA Time Zone database, such as `America/New_York`

The time zone database needed by LoadLocation may not be present on all systems, especially non-Unix systems. The rules looks in the directory or uncompressed zip file
named by the `ZONEINFO` environment variable, if any, then looks in known installation locations on Unix systems, and finally looks in `$GOROOT/lib/time/zoneinfo.zip`.

#### ip

The field under validation must be a string and be either an IPv4 address or an IPv6 address.
This rule converts the field to `net.IP` if it passes.

#### ipv4

The field under validation must be a string and be an IPv4 address.
This rule converts the field to `net.IP` if it passes.

#### ipv6

The field under validation must be a string and be an IPv6 address.
This rule converts the field to `net.IP` if it passes.

#### json

The field under validation must be a valid JSON string. This rule unmarshals the string and sets the field value to the unmarshalled result.

#### url

The field under validation must be a valid URL.
This rule converts the field to `*url.URL` if it passes.

#### uuid:version

The field under validation must be a string and a valid UUID.

The `version` parameter is **optional**.
- If a `version` is given (`uuid:3`,`uuid:4`,`uuid:5`), the rule will pass only if the version of the UUID matches.  
- If no `version` is given, any UUID version is accepted.

This rule converts the field to `uuid.UUID` if it passes.

#### bool

The field under validation must be a boolean or one of the following values:
- `1`/`0`
- `"1"`/`"0"`
- `"on"`/`"off"`
- `"true"`/`"false"`
- `"yes"`/`"no"`

This rule converts the field to `bool` if it passes.

#### same:field

The field under validation must have the same value as the given `field`. For arrays, the two fields must have the same values in the same order.

The two fields must have the same type. Files are not checked.

#### different:field

The field under validation must have a different value from the given `field`. For arrays, the two fields must have different values or not be in the same order.

The two fields must have the same type. Files are not checked.

#### confirmed

The field under validation must have a matching `foo_confirmation`. This rule validate equality in the same way as the [`same`](#same-field) rule.

For example, if the field under validation is `password`, a matching `password_confirmation` field must be present in the input.

#### file

The field under validation must be a file. Multi-files are supported.
This rule converts the field to `[]filesystem.File` if it passes.

#### mime:foo,...

The field under validation must be a file and match one of the given MIME types. If the field is a multi-files, all files must satisfy this rule.

#### image

The field under validation must be a file and match one of the following MIME types:
- `image/jpeg`
- `image/png`
- `image/gif`
- `image/bmp`
- `image/svg+xml`
- `image/webp`

If the field is a multi-files, all files must satisfy this rule.

#### extension:foo,...

The field under validation must be a file and match one of the given extensions. If the field is a multi-files, all files must satisfy this rule.

#### count:value

The field under validation must be a multi-file and contain `value` files.

#### count_min:value

The field under validation must be a multi-file and contain at least `value` files.

#### count_max:value

The field under validation must be a multi-file and may not contain more than `value` files.

#### count_between:min,max

The field under validation must be a multi-file and contain between `min` and `max` files.

#### date:format

The field under validation must be a string representing a date. The `format` is optional. If no format is given, the `2006-01-02` format is used.

This rule converts the field to `time.Time` if it passes.

::: tip
See the [Golang datetime format](https://golang.org/src/time/format.go).
:::

::: warning
When validating dates by comparing them together, the order of the declaration of the fields in the request is important. For example, if you want to validate that an end date is after a start date, the start date should be decalred **before** the end date in the rules set.

If a date has not been validated and converted yet, the date comparison rules will attempt to parse the dates using the following format: `2006-01-02`.
:::

#### before:date

The field under validation must be a value preceding the given date. The `date` must be written using following format: `2006-01-02T15:04:05`.

If the name of another field is given as a `date`, then the two fields must be a date and the field under validation must be preceding the given field.

#### before_equal:date

The field under validation must be a value preceding or equal to the given date. The `date` must be written using following format: `2006-01-02T15:04:05`.

If the name of another field is given as a `date`, then the two fields must be a date and the field under validation must be preceding or equal to given field.

#### after:date

The field under validation must be a value after the given date. The `date` must be written using following format: `2006-01-02T15:04:05`.

If the name of another field is given as a `date`, then the two fields must be a date and the field under validation must be preceding the given field.

#### after_equal:date

The field under validation must be a value after or equal to the given date. The `date` must be written using following format: `2006-01-02T15:04:05`.

If the name of another field is given as a `date`, then the two fields must be a date and the field under validation must be after or equal to given field.

#### date_equals:date

The field under validation must be a value equal to the given date. The `date` must be written using following format: `2006-01-02T15:04:05`.

If the name of another field is given as a `date`, then the two fields must be a date and the field under validation must be equal to given field.

#### date_between:date1,date2

The field under validation must be a value between or equal to the given dates. The given dates must be written using following format: `2006-01-02T15:04:05`.

If the name of another field is given as a date, then all the fields must be a date and the field under validation must be between or equal to given fields.

## Custom rules

If none of the available validation rules satisfy your needs, you can implement custom validation rules. To do so, create a new file `http/requests/validation.go` in which you are going to define your custom rules.

Rules definition shouldn't be exported, and start with `validate`. A rule returns a `bool`, indicating if the validation passed or not.

``` go
func validateCustomFormat(field string, value interface{}, parameters []string, form map[string]interface{}) bool {
    // Ensure the rule has at least one parameter
    validation.RequireParametersCount("custom_format", parameters, 1)
    str, ok := value.(string)

    if ok { // The data under validation is a string
        return regexp.MustCompile(parameters[0]).MatchString(str)
    }

    return false // Cannot validate this field
}
```
::: tip
- `validation.Rule` is an alias for `func(string, interface{}, []string, map[string]interface{}) bool`
- The `form` parameter lets you access the whole form data, and modify it if needed.
- The custom rule in the example above validates a string using a regex. If you need this kind of validation, prefer the included `regex` validation rule.
:::

Now that your rule behavior is defined, you need to **register** your rule. Do this in an `init()` function in your `validation.go` file.

#### validation.AddRule

Register a validation rule.

The rule will be usable in request validation by using the given rule name.

Type-dependent messages let you define a different message for numeric, string, arrays and files. The language entry used will be "validation.rules.rulename.type"

| Parameters                  | Return |
|-----------------------------|--------|
| `name string`               | `void` |
| `typeDependentMessage bool` |        |
| `rule validation.Rule`      |        |

**Example:**
``` go
func init() {
    validation.AddRule("custom_format", false, validateCustomFormat)
}
```

#### validation.RequireParametersCount

Checks if the given parameters slice has at least `count` elements. If this criteria is not met, the function triggers a panic.

Use this to make sure your validation rules are correctly used.

| Parameters        | Return |
|-------------------|--------|
| `rule string`     | `void` |
| `params []string` |        |
| `count int`       |        |

**Example:**
``` go
validation.RequireParametersCount("custom_format", parameters, 1)
```

#### validation.GetFieldType

<p><Badge text="Since v2.0.0"/></p>

returns the non-technical type of the given `value` interface.
This is used by validation rules to know if the input data is a candidate
for validation or not and is especially useful for type-dependent rules.
- `numeric` if the value is an int, uint or a float
- `string` if the value is a string
- `array` if the value is a slice
- `file` if the value is a slice of `filesystem.File`
- `unsupported` otherwise

| Parameters          | Return   |
|---------------------|----------|
| `value interface{}` | `string` |

**Example:**
``` go
validation.GetFieldType("foo") // "string"
validation.GetFieldType(2) // "numeric"
validation.GetFieldType(2.4) // "numeric"
validation.GetFieldType([]int{1,2}) // "array"
```

## Validating arrays

<p><Badge text="Since v2.1.0"/><Badge text="BETA" type="warn"/></p>

Validating arrays is easy. All the validation rules, **except the file-related rules and the `confirmed` rule**, can be applied to array values using the prefix `>`. When array values are validated, **all of them** must pass the validation.

**Example:**
``` go
var arrayValidation = goyave.RuleSet{
    "array": {"required", "array:string", "between:1,5", ">email", ">max:128"},
}
```
In this example, we are validating an array of one to five email addresses, which can't be longer than 128 characters.

### N-dimensional arrays

You can validate n-dimensional arrays. 

**Example:**
``` go
var arrayValidation = RuleSet{
    "array": {"required", "array", ">array", ">>array:numeric", ">max:3", ">>>max:4"},
}
```
In this example, we are validating a three-dimensional array of numeric values. The second dimension must be made of arrays with a size of 3 or less. The third dimensions must contain numbers inferior to 4. The following JSON input passes the validation:
```json
{
    "array": [
        [[0.5, 1.42], [0.6, 4, 3]],
        [[0.6, 1.43], [], [2]]
    ]
}
```

## Placeholders

Validation messages can use placeholders to inject dynamic values in the validation error message. For example, in the `rules.json` language file:

```json
"between.string": "The :field must be between :min and :max characters."
```

Here, the `:field` placeholder will be replaced by the field name, `:min`  with the first parameter and `:max` with the second parameter, effectively giving the following result:

```
The password must be between 6 and 32 characters. 
```

Placeholders are **replacer functions**. In fact, `validation.Placeholder` is an alias for `func(string, string, []string, string) string`. These functions should return the value to replace the placeholder with.

**Example:**
``` go
func simpleParameterPlaceholder(field string, rule string, parameters []string, language string) string {
	return parameters[0]
}
```

---

Placeholders are implemented in the `http/requests/placeholders.go`. To register a custom placeholder, use the `validation.SetPlaceholder()` function, preferably in the `init()` function of your `placeholders.go` file.

#### validation.SetPlaceholder

Sets the replacer function for the given placeholder. Don't include the colon prefix in the placeholder name.

If a placeholder with this name already exists, the latter will be overridden.


| Parameters                        | Return |
|-----------------------------------|--------|
| `placeholderName string`          | `void` |
| `replacer validation.Placeholder` |        |

**Example:**
``` go
validation.SetPlaceholder("min", func(field string, rule string, parameters []string, language string) string {
  	return parameters[0] // Replace ":min" by the first parameter in the rule definition
})
```

### Available placeholders

#### :field

`:field` is replaced by the name of the field. If it exists, the replacer with favor the language lines in `fields.json`.

#### :other

`:other` is replaced by the name of the field given as first parameter in the rule definition. If it exists, the replacer with favor the language lines in `fields.json`.

For example, the `same:password_confirmation` rule compares two fields together and returns the following message if the validation fails: 
```
The :field and the :other must match.

The password and the password confirmation must match.
```

#### :value

`:value` is replaced by the first parameter of the rule definition.

#### :values

`:values` is replaced by a concatenation of all rule parameters, joined by a comma.

#### :min

`:min` is replaced by the first parameter of the rule definition.

#### :max

`:max` is replaced by the first parameter of the rule definition, or the second if the rule name contains `between`.

#### :version

`:version` is replaced by a concatenation of `v` and the first parameter of the rule definition, or by an empty string if the rule doesn't have any parameter.

For example, for the `UUID:4` rule, the result would be `v4`.

#### :date

`:date` is replaced by the first parameter of the rule definition. If the first parameter is a field name, `:date` will be replaced with the name of the field in the same way as the `:other` placeholder.

#### :max_date

`:max_date` is replaced by the second parameter of the rule definition. If the second parameter is a field name, `:max_date` will be replaced with the name of the field in the same way as the `:other` placeholder.

## Manual validation

<p><Badge text="Since v2.1.0"/></p>

You may need to validate some data manually as part of your business logic. You can use the Goyave validator to do so.

#### validation.Validate

Validate the given data with the given rule set. If all validation rules pass, returns an empty `validation.Errors`.

The third parameter (`isJSON`) tells the function if the data comes from a JSON request. This is used to return the correct message if the given data is `nil` and to correctly handle arrays in url-encoded requests.

The last parameter (`language`) sets the language of the validation error messages.

| Parameters                    | Return              |
|-------------------------------|---------------------|
| `data map[string]interface{}` | `validation.Errors` |
| `rules RuleSet`               |                     |
| `isJSON bool`                 |                     |
| `language string`             |                     |

::: tip
`validation.Errors` is an alias for `map[string][]string`. The key represents the field name and the associated slice contains all already translated validation error messages for this field.
:::

**Example:**
``` go
func Store(response *goyave.Response, request *goyave.Request) {
    data := map[string]interface{}{
		"string": "hello world",
		"number": 42,
	}

	errors := validation.Validate(data, validation.RuleSet{
		"string": {"required", "string"},
		"number": {"required", "numeric", "min:10"},
	}, true, request.Lang)

	if len(errors) > 0 {
		response.JSON(http.StatusUnprocessableEntity, map[string]validation.Errors{"validationError": errors})
		return
	}

	// data can be safely used from here
	// ...
}
```