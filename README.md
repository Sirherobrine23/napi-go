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

	napiStruct, err := napi.ValueOf(env, toGoReflect)
	if err != nil {
		panic(err)
	}
	export.Set("goStruct", napiStruct)

	fnCall, err := napi.GoFuncOf(env, func(call ...any) (string, error) {
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

	fnCallStruct, err := napi.GoFuncOf(env, func(call ...Test) (string, error) {
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
- [x] Thread safe function (explicit native queue)
- [x] Array buffer
- [x] Dataview
- [x] Typed Array
- [x] Class (explicit constructor and wrapped Go state)

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
- [x] Class (explicit constructor and wrapped Go state)

## Classes

Use `CreateClass(env, name, constructor, properties...)` to define a native
JavaScript constructor. Constructor, method, getter and setter callbacks all use
`*CallbackInfo`. Initialize `ci.This` in the constructor and return `nil` to use
that instance. The constructor must be invoked with `new`.

`PropertyDescriptor` supports `Utf8name` or a string/symbol `Name`, `Method`,
`Getter`, `Setter`, `Value`, and `Attributes`. Combine `Static` with the desired
attributes to place a property on the constructor. Other properties belong to
the prototype. Attributes default to `Default` (zero); use `DefaultMethod` for
writable/configurable methods or `DefaultJSProperty` for ordinary data properties.
`Object.DefineProperties` uses the same descriptors for an existing object.

Attach Go state with `Object.Wrap(data, finalize)`, retrieve it with `Unwrap`,
and transfer ownership back to Go with `RemoveWrap`. A successful removal cancels
the finalizer. Wraps use `runtime/cgo.Handle` with C-allocated tokens and a native
type tag; other addons' native pointers are never interpreted as Go handles.
The finalizer runs on GC or environment teardown, must not call Node-API or
JavaScript, and receives the Go state rather than an environment handle.

`Class.New(args...)` constructs an instance from Go, including with zero args.
`ToClass` / `As[*Class]` wrap an existing JavaScript constructor. Explicit class
APIs work in both directions; reflection-based Go struct conversion still
produces plain JavaScript objects and does not automatically expose Go methods.

Like all `ValueType` objects, a `Class` is tied to its Node thread, environment
and handle scope. Do not retain its raw handle across callbacks without a native
reference. Class/descriptor callbacks are retained until their environment exits
so that extracted methods and accessors remain valid after constructor GC.
Wrapped instance state is independently released by GC.

See [the Counter example](internal/examples/class/main.go):

```sh
go build -buildmode=c-shared -o counter.node ./internal/examples/class
node -e 'const {Counter} = require("./counter.node"); console.log(new Counter(23).increment())'
```

## Thread-safe functions

Create a `ThreadsafeFunction` on the Node thread, then send Go-owned data from
goroutines with `Call`. `CreateThreadsafeFunction` accepts a JavaScript
`*Function`/`ValueType`, a Go `Callback`, an unnamed function with the same callback
signature, or `nil` when a custom `ThreadsafeFunctionCallJSCallback` is supplied.
With a nil custom callback, Go data is converted on the Node thread and delivered
to the JavaScript function as one argument. Nil data calls the function with no
arguments. A custom callback receives `env`, the optional JavaScript function,
and the original Go data; it can choose its own argument conversion.

The parameter order remains:

```go
fn, err := napi.CreateThreadsafeFunction(
    env, jsFunction, finalizeCallback, callJSCallback,
    "resourceName", maxQueueSize, initialThreadCount, context,
)
```

- Queue size `0` is unbounded; negative sizes and thread counts below `1` fail.
- Each producer owns an initial acquisition or calls `Acquire` before starting,
  then calls `Release(napi.Release)` once when finished. A release does not
  invalidate other producers. Pending calls drain before native finalization.
- `Call(data, napi.NonBlocking)` returns a queue-full error without taking
  ownership of the data when the bounded queue is full. Use `errors.Is` with
  `ErrThreadsafeFunctionQueueFull` / `ErrThreadsafeFunctionClosing`.
- `Call(data, napi.Blocking)` waits for queue space. Use it only from background
  goroutines; calling it on the JavaScript thread can deadlock the event loop.
- `Release(napi.Abort)` closes the queue and wakes blocked producers. Other
  producers must stop and release their own acquisitions (or receive the closing
  error if native finalization already retired them). Abandoned queued data
  is freed even when Node cannot invoke JavaScript during shutdown.
- `Ref(env)` / `Unref(env)` control event-loop liveness. They are idempotent,
  must run on the Node thread, and do not change producer acquisitions.
- `GetContext` works from goroutines and fails after native finalization. The
  optional finalizer runs once; internal handles are also freed without one.
- Send Go-owned values, not Node `ValueType` handles (including inside containers).
  Keep data immutable until delivery or synchronize shared access yourself.

A `ThreadsafeFunction` is a native resource and no longer implements `ValueType`.
Export a JavaScript entry function that starts work; do not export the queue itself.
See [the goroutine example](internal/examples/threadsafe/main.go):

```sh
go build -buildmode=c-shared -o threadsafe.node ./internal/examples/threadsafe
node -e 'require("./threadsafe.node").start(console.log)'
```

## Validation

Go 1.24 or later, a C compiler, Node.js and matching Node development headers
are required. Install your distribution's Node development package, or point
`CGO_CFLAGS` at a directory containing `node/node_api.h`. The integration suite
builds a real `.node` addon and exercises JavaScript, native GC and Worker teardown.

```sh
go build ./...
go test ./...
GOEXPERIMENT=cgocheck2 go test ./internal/integration
```
