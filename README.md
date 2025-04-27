# napi-go

A Go library for building Node.js Native Addons using Node-API.

## Usage

Get module with `go get -u sirherobrine23.com.br/Sirherobrine23/napi-go@latest`

register function to export values on module start

```go
package main

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/entry"
)

func init() {
	entry.Register(func(env napi.EnvType, export *napi.Object) {
		fromGoValue, _ := napi.CreateString(env, "hello from golang")
		export.Set("test", fromGoValue)
	})
}

func main() {}
```

Finally, build the Node.js addon using `go build`:

```sh
go build -buildmode=c-shared -o "example.node" .
```

## Go bind

Now it's easier to convert types from go to Javascript, current conversions:

- [x] Function
- [x] Struct, Map
- [x] Slice and Array
- [x] String
- [x] Int*, Uint* and Float
- [x] Boolean
- [x] Interface (interface if is `any` and not nil)
- [ ] Promise
  - [ ] Async Worker
  - [ ] Thread safe function
- [ ] Typed Array
  - [ ] Array buffer
    - [ ] Dataview

and convert Javascript values to Go values

- [x] Struct, Map
- [x] Slice and Array
- [x] String
- [x] Int*, Uint* and Float
- [x] Boolean
- [x] Interface (if is `any` set types map[string]any, []any or primitive values)
- [ ] Function
- [ ] Promise
- [ ] Typed Array
  - [ ] Array buffer
    - [ ] Dataview