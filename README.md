# tablemap

A Go library for easily converting between tabular data and Go structs with flexible marshaling options.

## Features

- Convert between table data (headers and rows) and Go structs with field tags to map table columns
- Support for custom marshaling/unmarshaling through interfaces
- Handle nil values with configurable representation

## Installation

```bash
go get github.com/kmio11/tablemap
```

## Basic Usage

### Struct Tags

The library uses `table` struct tags to map struct fields to table columns. The tag value specifies the column name in the table data:

```go
type Person struct {
    Name      string `table:"name"`       // Maps to "name" column
    Age       int    `table:"age"`        // Maps to "age" column
    Email     string `table:"email"`      // Maps to "email" column
    CreatedAt time.Time                   // Ignored (no tag)
}
```

- Fields with a `table` tag are mapped to columns with the specified name
- Fields without a `table` tag are ignored during marshaling/unmarshaling

### Marshal/Unmarshal

```go
// Marshal structs to table data
persons := []Person{
    {Name: "John Doe", Age: 30, Email: "john@example.com"},
    {Name: "Jane Smith", Age: 25, Email: "jane@example.com"},
}

header, data, err := table.Marshal(persons)
if err != nil {
    panic(err)
}

// Unmarshal table data to structs
var result []Person
err = table.Unmarshal(header, data, &result)
if err != nil {
    panic(err)
}
```

For more examples, see [example_test.go](example_test.go)

## Custom Marshaling

The library supports two ways to implement custom marshaling:

1. Implement `TableMarshaler` and `TableUnmarshaler` interfaces:

    ```go
    type CustomType struct {
        value string
    }

    func (c *CustomType) MarshalTable() (string, error) {
        return "custom:" + c.value, nil
    }

    func (c *CustomType) UnmarshalTable(s string) error {
        if len(s) > 7 && s[0:7] == "custom:" {
            c.value = s[7:]
        }
        return nil
    }
    ```

2. Implement `encoding.TextMarshaler` and `encoding.TextUnmarshaler` interfaces:

    If a type implements these standard Go interfaces, the library will automatically use them for marshaling and unmarshaling when `TableMarshaler`/`TableUnmarshaler` are not implemented.

## Options

Configure marshaling/unmarshaling behavior with `Options`.

### Handling Nil Values

The library provides flexible handling of nil values:

1. Default behavior:
```go
type Record struct {
    Name    *string `table:"name"`    // Will be "\N" when nil
    Age     *int    `table:"age"`     // Will be "\N" when nil
    Address *string `table:"address"` // Will be "\N" when nil
}
```

2. Custom nil representation:
```go
opts := &table.Options{
    NilValue: "NULL",
}

// Now nil values will be represented as "NULL" in the table data
header, data, err := table.MarshalWithOptions(records, opts)

// Unmarshaling will handle "NULL" values as nil
var result []Record
err = table.UnmarshalWithOptions(header, data, &result, opts)
```

## CSV Support

The `csvmap` package provides integration with CSV files.
See [csvmap/example_test.go](csvmap/example_test.go)

## License

MIT License - see [LICENSE](LICENSE) for details
