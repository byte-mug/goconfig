# goconfig
Yet another configuration Language

It is inspired by the configuration format seen in the [Current INN Documentation](https://www.eyrie.org/~eagle/software/inn/docs/).

[![GoDoc](https://godoc.org/github.com/byte-mug/goconfig?status.svg)](https://godoc.org/github.com/byte-mug/goconfig)

## Format Description

The configuration file is a fairly free-format file that consists of two types of entries: key-value pairs and groups.
Comments are from the hash character \# to the end of the line.

### Key-value pairs

Key-value pairs are a keyword and a value separated by a colon (which can itself be surrounded by whitespace). For example:
```
max-connections: 10
```
A legal key starts with a letter and contains only letters, digits, and the _ and - characters.
A legal value is any other sequence of characters. If the value needs to contain whitespace, then it must be quoted with double quotes (or single quotes as well), and uses the same format for embedding non-printing characters as normal C-language (or Go-language) string.

There are various different types of values such as integers, floating-point numbers, booleans, or strings. In the config file, they are all represented and parsed as strings and then converted into their corresponding type after parsing.

### Group entries

```
<class> <name> {
    # body ...
}
```

The `\<class\>` is any string valid as a key and is required. The `\<name\>` is any string valid as a value and can be omitted. The body of a group entry contains any number of the two types of entries.

## Mapping to structs.

In structs the key-value pairs and groups are assigned as as Keys. For this, annotations are used.

There are three sigils, `$`, `@` and `%`. In addition, the field tag has an optional suffix, `!`.
```go
struct {
	Field1 SomeType            `inn:"$field1"`  // A field has a sigil at it's start. For single values, we use '$'
	Field2 SomeType            `inn:"field2"`   // If no sigil is provided, '$' is default
	Field3 []SomeType          `inn:"@field3"`  // A field with the '@' sigil is an array.
	Field4 map[string]SomeType `inn:"%field4"`  // A field with the '%' sigil is a map.
	Field5 map[string]SomeType `inn:"%field5!"` // This field has the suffix '!'
}
```

The compatibility of between sigils and entry types.
| |key-value pairs|group entries|
|-|-|-|
|`$`|yes|yes|
|`@`|yes|yes|
|`%`|no|yes|

Key-Value pairs are handled as follows:

* The key is used to select the target field.
	* If the sigil of the target field is `$`, the value will be overwritten.
	* If the sigil of the target field is `@`, the value will be appended to the array.

Group entries are handled as follows:

* The `\<class\>` is used to select the target field.
	* If the sigil of the target field is `$`, the value will be overwritten.
	* If the sigil of the target field is `@`, the value will be appended to the array.
	* If the sigil of the target field is `%`, the value will be inserted into the map, with the `\<name\>` as key,
* Within the target-type, a Key-value pair will be inserted with `\<class\>` as key and `\<name\>` as value, provided that
	* The `\<name\>` is not omitted and non-empty
	* The target field has no `!`-suffix.

## Usage Example

```go
const code = `
    method tradspool {
        class: 1
        newsgroups: internal.*
    }
    method cnfs {
        class: 2
        newsgroups: alt.binaries.*
        options: BINARIES
    }
    method cnfs {
        class: 3
        newsgroups: *
        size: 50000
        options: LARGE
    }
    method timehash {
        class: 4
        newsgroups: alt.*
    }
    method timehash {
        class: 5
        newsgroups: *
    }
`

type StorageCfg struct {
	Method     string `inn:"$method"`
	Class      uint8  `inn:"$class"`
	Newsgroups string `inn:"$newsgroups"`
	Size       int    `inn:"$size"`
	Expires    string `inn:"$expires"`
	Options    string `inn:"$options"`
	Exactmatch bool   `inn:"$exactmatch"`
}
type Container struct {
	Storage []StorageCfg `inn:"@method"`
}

obj := new(Container)

err := goconfig.Parse([]byte(code),goconfig.CreateReflectHandler(obj))

if err!=nil { /* handle parse error */ }

/* process Container */
```


