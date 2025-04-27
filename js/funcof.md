# Go functions conversion to Javascript functions

## Simples conversions

| Go                                | Node                                        |
| --------------------------------- | ------------------------------------------- |
| `func()`                          | `function(): void`                          |
| `func() value`                    | `function(): value`                         |
| `func(value1, value2)`            | `function(value1, value2): void`            |
| `func() (value1, value2)`         | `function(): [value1, value2]`              |
| `func(value1, value2, values...)` | `function(value1, value2, ...values): void` |
| `func(values...)`                 | `function(...values): void`                 |

## Error throw

| Go                                         | Node                                                |
| ------------------------------------------ | --------------------------------------------------- |
| `func() error`                             | `function(): throw Error`                           |
| `func(value1, value2) error`               | `function(value1, value2): throw Error`             |
| `func() (value, error)`                    | `function(): value \|\| throw Error`                |
| `func(value1, value2) (...nValues, error)` | `function(value1, value2): (Array \|\|throw Error)` |
