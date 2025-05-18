# napi-go

A Go library for building Node.js Native Addons using Node-API.

>[!IMPORTANT]
>
> This module require [`CGO`](https://pkg.go.dev/cmd/cgo), to cross-compile check cgo documentation.
>
> My recommendation is to use a Linux distribution with the GCC installed together with the Required Destinations,
> or use [`zig`](https://ziglang.org/download/)(C: `zig cc -target <target>`, C++: `zig c++ -target <target>`)

## Usage

Get module with `go get -u sirherobrine23.com.br/Sirherobrine23/napi-go@latest`

register function to export values on module start, example

```go
package main

import (
	"encoding/json"
	"net/netip"
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/js"
)

type Test struct {
	Int    int
	String string
	Sub    []any
}

//go:linkname RegisterNapi sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func RegisterNapi(env napi.EnvType, export *napi.Object) {
	inNode, _ := napi.CreateString(env, "from golang napi string")
	inNode2, _ := napi.CopyBuffer(env, []byte{1, 0, 244, 21})
	toGoReflect := &Test{
		Int:    14,
		String: "From golang",
		Sub: []any{
			1,
			[]string{"test", "gopher"},
			[]bool{false, true},
			[]int{23, 244, 10, 2024, 2025, 2000},
			map[string]string{"exampleMap": "test"},
			map[int]string{1: "one"},
			map[bool]string{false: "false", true: "true"},
			map[[2]string]string{{"go"}: "example"},
			netip.IPv4Unspecified(),
			netip.IPv6Unspecified(),
			netip.AddrPortFrom(netip.IPv6Unspecified(), 19132),
			nil,
			true,
			false,
			inNode,
			inNode2,
			func() {
				println("called in go")
			},
		},
	}

	napiStruct, err := js.ValueOf(env, toGoReflect)
	if err != nil {
		panic(err)
	}
	export.Set("goStruct", napiStruct)

	fnCall, err := js.GoFuncOf(env, func(call ...any) (string, error) {
		d, err := json.MarshalIndent(call, "", "  ")
		if err == nil {
			println(string(d))
		}
		return string(d), err
	})
	if err != nil {
		panic(err)
	}
	export.Set("printAny", fnCall)

	fnCallStruct, err := js.GoFuncOf(env, func(call ...Test) (string, error) {
		d, err := json.MarshalIndent(call, "", "  ")
		if err == nil {
			println(string(d))
		}
		return string(d), err
	})
	if err != nil {
		panic(err)
	}
	export.Set("printTestStruct", fnCallStruct)
}

func main() {}
```

Finally, build the Node.js addon using `go build`:

```sh
go build -buildmode=c-shared -o "example.node" .
```

se more examples in [internal/examples](internal/examples)

## Go bind

Now there are some new functions that convert values ​​from Golang to JavaScript and vice-versa.

They are the:

- `ValueOf`: Convert go values to Javascript values.
- `ValueFrom` Convert Javascript values to Golang values.

### From Go to Javascript:

- [x] Function
- [x] Struct, Map
- [x] Slice and Array
- [x] String
- [x] Int*, Uint* and Float
- [x] Boolean
- [x] Interface (interface if is `any` and not nil)
- [x] Promise
  - [x] Async Worker
  - [ ] Thread safe function
- [x] Array buffer
- [x] Dataview
- [x] Typed Array
- [ ] Class

### Convert from Javascript values to Go

- [x] Struct, Map
- [x] Slice and Array
- [x] String
- [x] Int*, Uint* and Float
- [x] Boolean
- [x] Interface (if is `any` set types map[string]any, []any or primitive values)
- [x] Array buffer
- [x] Typed Array
- [x] Dataview
- [ ] Function
- [ ] Class
