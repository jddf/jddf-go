# jddf-go [![badge]][docs]

> Documentation on godoc.org: https://godoc.org/github.com/jddf/jddf-go

This package is a Golang implementation of **JSON Schema Langugage**. You can
use this package to:

1. Validate input data against a schema,
2. Get a list of validation errors from that input data, or
3. Build your own tooling on top of JSON Data Definition Format

## Usage

See [the docs][docs] for more detailed usage, but at a high level, here's how
you parse schemas and validate input data against them:

```golang
import (
  "encoding/json"
  "fmt"

  // This repository exports a package named "jddf".
  "github.com/jddf/jddf-go"
)

func main() error {
  // jddf.Schema can be parsed from JSON directly, but you can also construct
  // instances using the literal syntax:
  schema := jddf.Schema{
    RequiredProperties: map[string]jddf.Schema{
      "name":   jddf.Schema{Type: jddf.TypeString},
      "age":    jddf.Schema{Type: jddf.TypeUint32},
      "phones": jddf.Schema{
        Elements: &jddf.Schema{Type: jddf.TypeString}
      }
    }
  }

  // To keep this example simple, we'll construct this data by hand. But you
  // could also parse this data from JSON.
  //
  // This input data is perfect. It satisfies all the schema requirements.
  inputOk := map[string]interface{}{
    "name": "John Doe",
    "age":  43,
    "phones": [
      "+44 1234567",
      "+44 2345678",
    ],
  }

  // This input data has problems. "name" is missing, "age" has the wrong type,
  // and "phones[1]" has the wrong type.
  inputBad := map[string]interface{}{
    "age": "43",
    "phones": []interface{}{
      "+44 1234567",
      442345678,
    }
  }

  // To keep things simple, we'll ignore errors here. In this example, errors
  // are impossible. The docs explain in detail why an error might arise from
  // validation.
  validator := jddf.Validator{}
  resultOk, _ := validator.Validate(schema, inputOk)
  resultBad, _ := validator.Validate(schema, inputBad)

  fmt.Println(resultOk.IsValid()) // true
  fmt.Println(len(resultBad.Errors)) // 3

  // [] [properties name] -- indicates that the root is missing "name"
  fmt.Println(resultBad.Errors[0].InstancePath, resultBad.Errors[0].SchemaPath)

  // [age] [properties age type] -- indicates that "age" has the wrong type
  fmt.Println(resultBad.Errors[1].InstancePath, resultBad.Errors[1].SchemaPath)

  // [phones 1] [properties phones elements type] -- indicates that "phones[1]"
  // has the wrong type
  fmt.Println(resultBad.Errors[2].InstancePath, resultBad.Errors[2].SchemaPath)
}
```

[badge]: https://godoc.org/github.com/jddf/jddf-go?status.svg
[docs]: https://godoc.org/github.com/jddf/jddf-go
