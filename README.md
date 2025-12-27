# trongo

[![Go Reference](https://pkg.go.dev/badge/github.com/tron-format/trongo.svg)](https://pkg.go.dev/github.com/tron-format/trongo)
[![License: MIT](https://img.shields.io/badge/license-MIT-fef3c0?labelColor=1b1b1f)](./LICENSE)

A Go library for converting data to and from the TRON (Token Reduced Object Notation) format.

This library provides a Go implementation that matches the API of the standard `encoding/json` package, making it a drop-in replacement for JSON serialization with the benefits of TRON's token efficiency.

See full specification for the TRON format at: https://tron-format.github.io/

## Installation

```bash
go get github.com/tron-format/trongo
```

## Development

This project uses a Taskfile and depends on **go-task** (Task): https://taskfile.dev/

Common commands:

```bash
# list available tasks
task --list

# run the full local workflow (fmt/vet/build/test)
task all
```

## Usage

```go
package main

import (
    "fmt"
    "github.com/tron-format/trongo/pkg/tron"
)

func main() {
    type Person struct {
        Name string `json:"name"`
        Age  int    `json:"age"`
    }

    people := []Person{
        {Name: "Alice", Age: 30},
        {Name: "Bob", Age: 25},
    }

    // Marshal to TRON format
    data, err := tron.Marshal(people)
    if err != nil {
        panic(err)
    }
    fmt.Println(string(data))
    // Output:
    // class A: name,age
    //
    // [A("Alice",30),A("Bob",25)]

    // Unmarshal from TRON format
    var result []Person
    err = tron.Unmarshal(data, &result)
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", result)
    // Output: [{Name:Alice Age:30} {Name:Bob Age:25}]
}
```

## API Compatibility

This library provides the same API as Go's `encoding/json` package:

- `tron.Marshal(v interface{}) ([]byte, error)`
- `tron.Unmarshal(data []byte, v interface{}) error`
- `tron.MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)`
- Support for struct tags (`json:"fieldname"`)
- Support for custom `MarshalTRON()` and `UnmarshalTRON()` methods

## Features

- **Token Efficiency**: TRON format reduces redundancy by defining reusable class structures
- **JSON Compatibility**: Seamless migration from JSON with identical API
- **Type Safety**: Full Go type system support with struct tags
- **Performance**: Optimized for both encoding and decoding operations
- **Standards Compliant**: Follows the official TRON specification

## Playground

Want to try out TRON with your own data?

Go to https://tron-format.github.io/#/playground and select "Custom Data".

Paste in your data to see TRON's token efficiency compared to other data formats!

## License

[MIT](./LICENSE) License © 2025-PRESENT