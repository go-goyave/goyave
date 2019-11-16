# Validation

[[toc]]

## Introduction

Goyave provides a powerful, yet easy way to validate all incoming data, no matter its type or its format, thanks to a large number of validation rules.

Incoming requests are validated using **rules set**, which associate rules with each expected field in the request.

Validation rules can **alter the raw data**. That means that when you validate a field to be number, if the validation passes, you are ensured that the data you'll be using in your controller handler is a `float64`. Or if you're validating an IP, you get a `net.IP` object.

Validation is automatic. You just have to define a rules set and assign it to a route. When the validation doesn't pass, the request is stopped and the validation errors messages are sent as a response, using the correct [language](../advanced/localization). The HTTP response code of failed validation is **422 "Unprocessable Entity"**, or **400 "Bad Request"** if the body could not be parsed.

::: tip
You can customize the validation error messages by editing `resources/lang/<language>/rules.json`. Learn more in the [localization](../advanced/localization) section.
:::

## Rules sets

The `http/requests` directory contains the requests validation rules sets. You should have one package per feature, regrouping all requests handled by the same controller. The package should be named `<feature_name>Requests`.

**Example:** (`http/request/productRequests/product.go`)
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
router.Route("POST", "/product", productController.Store, productRequests.Store)
```

## Available validation rules

::: table
[Required](#required)
[Numeric](#numeric)
[Integer](#integer)
[Min](#min)
[Max](#max)
[Between](#between)
[Greater than](#greater_than)
[Greater than or equal](#greater_than_equal)
[Lower than](#lower_than)
[Lower than or equal](#lower_than_equal)
[String](#string)
[Array](#array)
[Distinct](#distinct)
[Digits](#digits)
[Regex](#regex)
[Email](#email)
[Size](#size)
[Alpha](#alpha)
[Alpha dash](#alpha_dash)
[Alpha num](#alpha_num)
[Starts with](#starts_with)
[Ends with](#ends_with)
[In](#in)
[Not in](#not_in)
[In array](#in_array)
[Not in array](#not_in_array)
[Timezone](#timezone)
[IP](#ip)
[IPv4](#ipv4)
[IPv6](#ipv6)
[JSON](#json)
[URL](#url)
[UUID](#uuid)
[Bool](#bool)
[Same](#same)
[Different](#different)
[Confirmed](#confirmed)
[Dile](#file)
[MIME](#mime)
[Image](#image)
[Extension](#extension)
[Count](#count)
[Count min](#count_min)
[Count max](#count_max)
[Count between](#count_between)
[Date](#date)
[Before](#before)
[Before equal](#before_equal)
[After](#after)
[After equal](#after_equal)
[Date equals](#date_equals)
[Date between](#date_between)
[Nullable](#nullable)
:::

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
