# Node.js builtins from Go

These packages call the real Node builtins. They use the host's platform rules,
OpenSSL, filesystem permissions and event loop, rather than Go replacements.
Existing `EventEmitter` / `EventEmitterThread` APIs remain available.

Every wrapper is a scoped `napi.ValueType`. Create/load it on the owning Node
thread, use it in that callback, or return its JS value to JavaScript. Load a
fresh Go wrapper inside a later callback; there is no Go global cache of module
handles. Workers have independent Node environments.

Loading uses `process.getBuiltinModule` (Node 20.16+ / 22.3+), with a legacy
fallback to an explicitly available global `require`. CommonJS and ESM both
work on modern Node. If a host has neither loader, supply a scoped module object
to `nodejs.Wrap`. Unavailable modules/functions produce an error at call time.

## Common interface

All module packages expose `Load(env)` and embed `*nodejs.Module`. Its methods
cover the rest of each builtin API, even when no typed helper is provided:

```go
module, err := nodejs.Load(ci.Env, "node:path") // "path" also works
if err != nil {
    return nil, err
}
joined, err := nodejs.CallAs[string](module, "join", "a", "b")
if err != nil {
    return nil, err
}
return napi.ValueOf(ci.Env, joined)
```

| Method/helper                                                                         | Use                                                                         |
| ------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `nodejs.Load(env, id)`                                                                | Load any available builtin, including submodules such as `node:fs/promises` |
| `nodejs.Wrap(value, id)`                                                              | Bind an already loaded JS object/namespace in its current scope             |
| `module.Call(name, args...)`                                                          | Invoke a method with Go values and the module as `this`                     |
| `module.CallInto(name, &result, args...)` / `nodejs.CallAs[T](module, name, args...)` | Invoke and decode a typed result                                            |
| `module.Get`, `GetInto`, `SetValue`                                                   | Access properties/constants using the embedded Object helpers               |
| `module.New(name, args...)`                                                           | Construct an exported JS class with Go values                               |
| `module.CallPromise(name, args...)`                                                   | Invoke and check a genuine JS Promise result                                |
| `nodejs.CallbackOf(env, function)`                                                    | Adapt a Go function/Callback or validate an existing JS function            |
| `nodejs.Binary(env, input)`                                                           | Copy Go `[]byte` to Buffer; preserve normal conversion for other inputs     |
| `nodejs.CopyBytes(value)`                                                             | Copy a native Buffer into Go-owned memory on the Node thread                |

`Load` is restricted to builtins. It does not import npm packages or await dynamic
ESM imports. The generic API can load modules such as `node:http`, `node:net`,
`node:child_process` or `node:worker_threads` without a separate typed package.
Their availability and operation still follow the running Node version and host.

## Typed packages

| Package                                  | Implemented conveniences                                                                                                                                     |
| ---------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [`path`](path/path.go)                   | Join, resolve, normalize, parse/format, basename/dirname/extname, relative/isAbsolute, POSIX/Windows variants, separators                                    |
| [`os`](os/os.go)                         | Architecture/platform/hostname, OS type/release, home/temp directories, CPU/user/network information, memory, uptime/load, parallelism                       |
| [`fs`](fs/fs.go)                         | Sync binary/text reads, write/append, exists/readdir/stat, mkdir/temp directory, rename/unlink/rm; async read/write/stat/mkdir/readdir/rm via `LoadPromises` |
| [`buffer`](buffer/buffer.go)             | Buffer.from/alloc/concat/compare/isBuffer/byteLength, encodings and copies                                                                                   |
| [`crypto`](crypto/crypto.go)             | Native Hash/Hmac update/digest, algorithms, random bytes/UUID and timingSafeEqual                                                                            |
| [`url`](url/url.go)                      | Native URL/URLSearchParams construction, URL properties, search parameter mutation/query, file URLs, domain ASCII/Unicode                                    |
| [`util`](util/util.go)                   | Format/inspect, deep strict equality, promisify/callbackify                                                                                                  |
| [`zlib`](zlib/zlib.go)                   | Sync gzip, deflate, raw deflate, Brotli compression/decompression and unzip                                                                                  |
| [`timers`](timers/timers.go)             | Timeout/interval/immediate, clear, ref/unref/hasRef/refresh; Promise timeout/immediate via `LoadPromises`                                                    |
| [`stream`](stream/README.md)             | Native stream methods, Go io.Reader/Writer/ReadWriter adapters, and Promise CopyTo/CopyFrom pipelines                                                        |
| [`dns`](dns/dns.go)                      | Promise lookup/resolve/resolve4/resolve6/reverse using `node:dns/promises`                                                                                   |
| [`EventEmitter`](EventEmitter/values.go) | Additional `EmitValues`, `OnFunc`, `OnceFunc` conveniences; background queue remains `EventEmitterThread`                                                    |

These are conveniences over stable/common operations, rather than a complete
typed mirror of every Node export. Options accept ordinary Go maps/structs or
existing JS values. Use the common interface for additional options, methods,
constructors or version-specific features.

## Asynchronous work

The [Go io bridge](stream/README.md) also runs Read/Write/Close in Go goroutines
while preserving native Node stream backpressure and lifetime. Existing streams
can pipe to/from Go I/O through `CopyTo`/`CopyFrom`, returning JS Promises.

`nodejs.Promise` wraps a JS-owned Promise and deliberately has no `napi_deferred`
or Resolve/Reject methods. It can be returned directly to JS for `await`, or
chained on the Node thread through `Then`, `Catch` and `Finally`. These methods
accept Go functions/Callbacks or existing JS functions. A literal nil handler
is passed as JS `undefined`: `Then(nil, nil)`, `Catch(nil)` and `Finally(nil)`
create new chains that preserve the original value or rejection. Non-nil
handlers must be functions. A returned Promise is adopted; errors reject the
chain. `Catch` calls the object's `catch` method directly, preserving overrides
and lookup exceptions; a missing method is an error.

```go
files, err := fs.LoadPromises(ci.Env)
if err != nil {
    return nil, err
}
promise, err := files.ReadFile(filename, "utf8")
if err != nil {
    return nil, err
}
return promise.Then(func(call *napi.CallbackInfo) (napi.ValueType, error) {
    text, err := napi.ArgAs[string](call, 0)
    if err != nil {
        return nil, err
    }
    // This callback has a fresh scope. Load modules here if needed.
    return napi.ValueOf(call.Env, strings.ToUpper(text))
}, nil)
```

Promise callbacks, stream listeners and timer callbacks execute on the Node
thread. They are not Go goroutines. For Go computation or Go I/O in the
background, use a goroutine with `EventEmitterThread`/TSFN, and convert data only
after its delivery to the Node thread. A copied `[]byte` can be sent to Go code;
a scoped Buffer/URL/module/function handle cannot.

Do not block the Node thread waiting on a Promise or channel fed by a JS callback.
The sync `fs`, `zlib` and `crypto` operations keep their native blocking behavior.
Timers use milliseconds as a JS Number (`float64`), with Node's clamping rules.
Timer Ref/Unref controls Node event-loop liveness, not Go reference ownership.
Return timer objects to JS or resolve fresh handles from a retained JS object
before manipulating them in a later Go callback.

`Stream.Write` returns native backpressure: false means the producer should wait
for `drain`. Do not spin or synchronously wait for that event on the Node thread.
Register stream `error` handlers and handle Promise rejections in the caller.

## Values and exceptions

Generic `Call` uses `napi.ValueOf`: `[]byte` is a JS array in the generic API.
Binary-specific helpers (fs write/append, crypto update/Hmac key, zlib transforms,
stream write/end) copy `[]byte` into a Buffer. `buffer.From` follows Buffer.from
and accepts the converted byte array. Existing JS Buffer handles pass through.
`CopyBytes` creates independent Go memory; `napi.Buffer.Data` is still a scoped
view of Node-owned memory. Numeric conversion keeps existing napi-go rules,
including Go `int64`/`uint64` -> JS BigInt and `int` -> JS Number.

JS getters, methods and async operations keep their original thrown/rejected
errors, including native codes such as ENOENT. Go Callbacks still convert returned
errors/panics into JS exceptions. Go synchronous callers receive the existing
Node-API status errors when the JS operation throws.

`URLSearchParams.Get` returns `(text, found, error)` so a missing key differs from
a present empty string. Hash instances use the native one-use digest contract;
using a finalized hash again reports the Node error. DNS Lookup uses the host
resolver, while Resolve methods perform DNS queries.

## Example and validation

The [example](../internal/examples/nodejs/main.go) reads a file asynchronously,
creates SHA-256 inside the Promise callback's scope, and returns its digest to JS.

```sh
make examples
node internal/examples/nodejs/index.cjs
make test-nodejs
make check
```

The 24 builtin tests exercise the real modules, native errors, binary copies,
empty data, sync/Promise filesystem calls, Promise adoption, timers, stream
backpressure/listeners, localhost DNS, CommonJS/ESM and Worker environments.
No test depends on public DNS or a remote server. Host network-interface access
is compared to direct Node calls, including permission failures in restricted
environments. The same suite participates in `make test`, `test-js` and
`test-cgo-check`. See [validation](../docs/nodejs-validation.md).

Primary API references: [Node builtins](https://nodejs.org/api/),
[process.getBuiltinModule](https://nodejs.org/api/process.html#processgetbuiltinmoduleid),
[fs](https://nodejs.org/api/fs.html), [crypto](https://nodejs.org/api/crypto.html),
[timers](https://nodejs.org/api/timers.html), [stream](https://nodejs.org/api/stream.html).
