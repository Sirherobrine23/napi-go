# napi-go

A Go library for building Node.js Native Addons using Node-API.

> [!IMPORTANT]
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

## Objects with Go values

The existing `CreateObject`, `Get`, `Set`, `Call` and `Class.New` APIs remain
available. The following helpers remove the intermediate `ValueOf` calls and
unchecked casts when working with Go data. They all require the owning Node
thread and a live handle scope, including when called from a TSFN callback.

```go
object, err := napi.ObjectOf(env, map[string]any{
    "name": "Gopher",
    "count": 23,
    "options": map[string]any{"enabled": true},
})
if err != nil {
    return nil, err
}
if err := object.SetValue("items", []int{1, 2, 3}); err != nil {
    return nil, err
}
name, err := napi.GetAs[string](object, "name")
if err != nil {
    return nil, err
}
if err := object.SetMethod("greet", func(who string) string {
    return "Hello " + who
}); err != nil {
    return nil, err
}
result, err := object.CallMethod("greet", name) // this == object
if err != nil {
    return nil, err
}
message, err := napi.Decode[string](result)
```

| API                                                        | Purpose                                                                        |
| ---------------------------------------------------------- | ------------------------------------------------------------------------------ |
| `ObjectOf(env, value)`                                     | Convert a map/struct or wrap an existing JS object; reject primitives and null |
| `object.SetValue(name, value)`                             | Assign one Go value or an existing `ValueType`                                 |
| `object.Assign(map[string]any{...})`                       | Assign multiple properties in sorted key order                                 |
| `object.GetInto(name, &target)` / `GetAs[T](object, name)` | Decode one property to a Go type or checked N-API wrapper                      |
| `object.SetMethod(name, function)`                         | Register a `Callback` or ordinary Go function                                  |
| `object.CallMethod(name, args...)`                         | Convert Go arguments and invoke the current method with the object as `this`   |
| `Decode[T](value)`                                         | Decode a standalone value using `ValueFrom`                                    |
| `ci.ArgInto(index, &target)` / `ArgAs[T](ci, index)`       | Decode an argument and report missing or incompatible values                   |
| `function.CallValues(args...)`                             | Call with Go arguments and the global object as `this`                         |
| `function.CallWithThis(receiver, args...)`                 | Call with Go arguments and an explicit receiver                                |
| `class.NewValues(args...)`                                 | Construct an instance using Go arguments                                       |
| `UnwrapAs[T](object)`                                      | Retrieve wrapped Go state with a checked type assertion                        |

Use an ordinary Go function for automatic argument/result conversion. Use a
`Callback` (or an unnamed function with the same signature) when you need the
environment, optional arguments, or `this`:

```go
err := object.SetMethod("increment", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
    amount, err := napi.ArgAs[int](ci, 0)
    if err != nil {
        return nil, err
    }
    this := napi.ToObject(ci.This)
    count, err := napi.GetAs[int](this, "count")
    if err != nil {
        return nil, err
    }
    if err := this.SetValue("count", count + amount); err != nil {
        return nil, err
    }
    return napi.ValueOf(ci.Env, count + amount)
})
```

Callbacks retrieve `ci.This` in their current scope; do not capture a scoped
`Object` handle from an earlier callback. Existing handles passed to conversions,
property assignment, calls or construction must belong to the same environment.
Background goroutines still send Go-owned data through `EventEmitterThread` or
a TSFN; the new object helpers do not make JavaScript handles thread-safe.

Conversion follows `ValueOf` / `ValueFrom`, including `napi:"field"` struct tags.
`int` becomes a JS Number; `int64` and `uint64` become BigInt. Bare Go `nil`, nil
pointers and nil functions become `undefined`; an explicit `env.Null()` value
preserves JS null. Nil maps/slices retain their existing empty-object/array
conversion. Reading JS null/undefined into a pointer, map or slice yields nil;
reading them into `any` yields nil. String-to-number coercion is not performed.
Decoded wrappers and raw handles inside `any` remain scoped to the Node thread.
Conversions have a recursion budget of 64 to report cyclic/excessively nested
data as an error.

Ordinary Go functions validate fixed argument counts and variadic minimums.
A trailing Go error becomes a JS exception; nil errors are omitted from results.
`SetMethod` rejects nil functions. Getter, setter, method and constructor
exceptions propagate to JavaScript. `Assign` can leave earlier properties changed
when a later conversion or setter fails. `SetValue` and `Assign` use the existing
`Object.Set` / Node-API semantics: a non-writable property may ignore assignment.
Use `DefineProperties` when you need explicit descriptor attributes.

See [the objects example](internal/examples/objects/main.go):

```sh
make examples
node internal/examples/objects/index.cjs
make test-objects
```

## Node.js builtin modules

The [`nodejs` packages](nodejs/README.md) wrap the real host modules through a
shared `Module` API and typed helpers for `path`, `os`, `fs`, `buffer`, `crypto`,
`url`, `util`, `zlib`, `timers`, `stream` and `dns`. `fs/promises` and
`timers/promises` return genuine JavaScript Promises with Go/JS
`Then`/`Catch`/`Finally` callbacks. Existing EventEmitter APIs also gain
`EmitValues`, `OnFunc` and `OnceFunc`.

All wrappers require the owning Node thread and current handle scope; load them
inside each callback. Background Go work continues to use TSFN/EventEmitterThread.
Sync filesystem/compression calls keep their native blocking behavior. Use the
generic `nodejs.Load(env, id)` / `Module.Call` API for additional builtin exports.

```sh
make examples
node internal/examples/nodejs/index.cjs
make test-nodejs
```

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

## Background tasks and EventEmitter

The [EventEmitter example](internal/examples/eventemitter) processes values in a
Go goroutine and sends `progress`, `done` and `error` packets through a bounded
TSFN queue. JavaScript relays packets to an `EventEmitter` on its owning Node
thread. Register listeners before calling `start`:

```js
const BackgroundTask = require("./internal/examples/eventemitter/index.cjs");
const task = new BackgroundTask();
task.on("progress", (data) => console.log(data));
task.once("done", (result) => console.log(result));
task.on("error", (error) => console.error(error));
task.start(10);
```

The queue retains the JavaScript callback until finalization, so the callback
can capture its emitter. Send immutable Go-owned data through `Call`; do not
call `Function.Call` or other JavaScript APIs directly from the goroutine.
`EventEmitter` listeners still run synchronously on the Node thread.

The integration suite additionally covers an existing
`emitter.emit.bind(emitter)` through a custom TSFN callback that expands the
event name and arguments. It verifies producer ordering, payloads, finalization,
empty tasks, error delivery, cancellation and Worker ownership.

## Makefile

```sh
make build                 # compile all Go packages
make examples              # build all included example addons
node internal/examples/eventemitter/index.cjs
make test                  # all packages and all Node integration suites
make test-eventemitter     # only the EventEmitter tests
make test-objects          # object helpers and Go conversion tests
make test-js               # all Node suites against build/integration.node
make test-cgo-check        # integration with GOEXPERIMENT=cgocheck2
make check                 # build + test + strict cgo checks
make test-race              # instrument the native addon with -race
make clean                 # remove the known generated addons/headers
```

For an official Node headers archive, `NODE_INCLUDE` must point at its `include`
directory, whose child is `node/node_api.h`; it must not point at `include/node`.
Existing `CGO_CPPFLAGS` and `CGO_CFLAGS` are preserved:

```sh
make NODE_INCLUDE=/path/to/node-headers/include check
make GO=/path/to/go NODE=/path/to/node BUILD_DIR=out examples test-js
```

`GO_TEST_FLAGS` defaults to `-count=1` so Makefile test targets execute the Node
suites instead of returning a cached result. For example, use
`make GO_TEST_FLAGS='-count=1 -timeout=4m' test-integration` for a slower builder.
Missing Node is reported as a failed check. `make vet` stays separate because
of the preexisting raw-finalizer warnings described in
[the validation report](docs/class-threadsafe-validation.md).

## Native EventEmitterThread

`nodejs/EventEmitter` now includes a native `EventEmitterThread`. Create it on
its owning Node thread, retrieve the JavaScript emitter in that scope, and send
Go-owned values from a goroutine. No JavaScript relay callback is required.

```go
thread, err := event.CreateEventThread(ci.Env, 64)
if err != nil {
    return nil, err
}
emitter, err := thread.Emitter(ci.Env)
if err != nil {
    thread.Abort()
    return nil, err
}
if err := thread.Run(func(ctx context.Context, events *event.EventEmitterThread) error {
    // Perform Go computation or I/O here; observe ctx cancellation as needed.
    result := 42
    return events.EmitBlocking("done", result)
}); err != nil {
    thread.Abort()
    return nil, err
}
return emitter, nil
```

Use the existing package path with an alias such as
`event "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/EventEmitter"`.
`Run` starts one processor goroutine, closes its queue when the processor returns,
and converts returned errors or recovered panics into JavaScript `Error` objects
on the `error` event. Register an error listener before the Node event loop starts
delivering events. Context cancellation and native queue closure stop the task
without generating an error event.

You can also wrap a JavaScript-provided emitter with
`event.ToEventEmitter(value).Thread(queueSize)`, or use `go func()` yourself and
call `Close` when all your producers finish. `Run` must wait for any child
producers it launches before returning. It cannot forcibly stop Go computation.

| Method                                     | Execution context                   | Behavior                                                        |
| ------------------------------------------ | ----------------------------------- | --------------------------------------------------------------- |
| `CreateEventThread`, `EventEmitter.Thread` | Owning Node thread                  | Create the emitter or wrap one, then create its queue           |
| `Emitter(env)`                             | Owning Node thread                  | Resolve a fresh JavaScript handle in the current callback scope |
| `Emit(name, args...)`                      | Go goroutines or owning Node thread | Enqueue without waiting; report queue-full/closing errors       |
| `EmitBlocking(name, args...)`              | Background Go goroutines            | Wait for bounded queue space                                    |
| `Run(processor)`                           | Go or owning Node thread            | Start one background processor and close on completion          |
| `Close()`                                  | Go or owning Node thread            | Cancel the processor context and drain accepted events          |
| `Abort()`                                  | Go or owning Node thread            | Cancel the processor context and discard pending events         |
| `Done()`                                   | Wait only in background Go code     | Signal native finalization and reference cleanup                |
| `Ref(env)`, `Unref(env)`                   | Owning Node thread                  | Control whether the queue keeps the event loop alive            |

`Emit` returns queue acceptance, not the synchronous boolean of JavaScript's
`emitter.emit`. The queue preserves enqueue order. With several producers their
events can interleave; each producer's sequential calls retain their order.
`Close` and `Abort` are idempotent; the first call selects the shutdown mode.
The emitter remains available to JavaScript after queue closure if JS retains it,
but `thread.Emitter` fails once native finalization removed its reference.

Keep payloads immutable until delivered (or until `Done`). Arguments are copied
as a slice, but their maps, pointers and nested containers are not deep-copied.
Scoped Node `ValueType` handles are rejected, including inside Go containers.
Cycles and payloads nested more than 64 levels are rejected before queuing.
Go `error` arguments become JavaScript `Error` objects on the Node thread.

The queue owns a persistent emitter reference and a bound `emit` callback,
released on native finalization or environment teardown. It is a native resource
and does not implement `ValueType`; export the scoped result of `Emitter(env)`.
The ordinary `EventEmitter` wrapper remains synchronous and requires the Node
thread; its constructor, event-name forwarding and method receivers are fixed.

The default constructor uses `process.getBuiltinModule` (Node 20.16+ / 22.3+),
which works in CommonJS, ESM and Workers without a global `require`. Older hosts
can supply a global `require` or pass an existing emitter to `.Thread`.
This uses `require.Builtin`; the generic asynchronous `require.Import` API is
not part of this event queue implementation.

```sh
make examples
node internal/examples/events_thread/index.cjs
make test-eventemitter-thread
make check
```

The [native example](internal/examples/events_thread/main.go) returns a real
EventEmitter with `progress`, `done`, `error` and a JavaScript `cancel()` method.
