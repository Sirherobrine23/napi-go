# Go io and Node streams

`stream.Load(env)` provides Go-backed binary Node streams and asynchronous
pipeline helpers. The objects are actual `Readable`, `Writable` and `Duplex`
instances, so JavaScript can use `pipe`, `pipeline`, async iteration, `finished`,
events and its ordinary stream APIs.

| API                                                | Direction                                                            |
| -------------------------------------------------- | -------------------------------------------------------------------- |
| `module.FromReader(io.Reader, options...)`         | Go Reader -> Node Readable                                           |
| `module.FromWriter(io.Writer, options...)`         | Node Writable -> Go Writer                                           |
| `module.FromDuplex(reader, writer, options...)`    | Separate Go endpoints -> Node Duplex                                 |
| `module.FromReadWriter(io.ReadWriter, options...)` | Shared Go endpoint -> Node Duplex                                    |
| `stream.CopyTo(io.Writer, options...)`             | Existing Node readable -> Go Writer via pipeline; returns JS Promise |
| `stream.CopyFrom(io.Reader, options...)`           | Go Reader -> existing Node writable via pipeline; returns JS Promise |

Constructors and Copy methods run on the owning Node thread in the current
callback scope. They do not perform the endpoint's Read/Write there. The bridge
starts Go goroutines for I/O and uses a TSFN to deliver results on Node's thread.
References retain each active stream/callback across scopes. Idle streams use
weak references and unref their TSFN, so they can be garbage-collected and do
not keep the process alive by themselves.

Existing `Stream.Read(size...)` / `Write(data...)` signatures remain available.
`*Stream` does not directly implement `io.Reader`/`io.Writer`: those methods have
different signatures, and blocking Go I/O must not be confused with scoped
Node-API operations. CopyTo/CopyFrom connect existing streams to Go interfaces
without introducing a blocking Go call on the Node thread.

## Reader example

Inside a native addon callback:

```go
module, err := stream.Load(ci.Env)
if err != nil {
    return nil, err
}
reader := io.MultiReader(strings.NewReader("hello "), strings.NewReader("from Go"))
return module.FromReader(reader, stream.IOOptions{ChunkSize: 4096})
```

The returned stream can be consumed normally:

```js
for await (const chunk of addon.read()) {
  console.log(chunk.toString());
}
```

Go reads start only when Node requests data. Each reader has at most one read
in flight; one chunk is pushed per request. Node controls subsequent requests
and applies its highWaterMark. A result with both bytes and `io.EOF` delivers
those bytes before pushing JS null. Empty EOF ends normally; repeated `(0, nil)`
reads fail after 100 attempts with `io.ErrNoProgress`. Other errors and panics
destroy the readable with a JavaScript Error.

## Writer and pipeline example

`hash.Hash`, `bytes.Buffer`, `bufio.Writer`, files, sockets and your own Go writers
can back a Writable. Node writes are copied into Go-owned memory before their
callback scope expires. They run serially; Node's write callback is acknowledged
only after the Go Write returns. A short nil-error write reports
`io.ErrShortWrite`. A writer with `Flush() error` is flushed before native finish.

```go
digest := sha256.New() // hash.Hash implements io.Writer
readable := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
promise, err := readable.CopyTo(digest)
if err != nil {
    return nil, err
}
return promise.Then(func(call *napi.CallbackInfo) (napi.ValueType, error) {
    return napi.CopyBuffer(call.Env, digest.Sum(nil))
}, nil)
```

```js
const digest = await addon.digest(Readable.from(["alpha", "beta"]));
```

`CopyTo` and `CopyFrom` use `node:stream/promises.pipeline`, including its error
propagation and participant destruction. Await the returned Promise in JS or
attach Node-thread callbacks. Native write backpressure and stream error handlers
keep their normal meaning; do not synchronously wait for callbacks on Node.

## Options, ownership and cancellation

`IOOptions` accepts a Context, ChunkSize, HighWaterMark and CloseOnDestroy. At
most one options value is accepted; zero sizes mean 32 KiB and negative sizes
are rejected. These adapters intentionally use binary mode, decodeStrings:true,
emitClose:true and autoDestroy:true.

By default, the caller retains ownership of Go endpoints. Set CloseOnDestroy
to transfer available io.Closer ownership after successful construction. Closers
run outside Node's thread and are called once, including errors, explicit
destroy, GC and Worker/environment teardown. A standalone owned writer is closed
before finish; an owned reader is closed when its Node readable is destroyed.
Close errors/panics become stream errors when JS delivery is still possible.

Context cancellation/deadline destroys the native stream with ABORT_ERR and
settles native callbacks. Generic Read/Write cannot be forcibly interrupted by
context alone. Owned Closers must allow Close concurrently with outstanding I/O
and unblock it, as supported by files, net.Conn and io.Pipe endpoints. For a
borrowed uninterruptible endpoint, its caller must unblock it. Native completion
does not wait for an uninterruptible borrowed operation; late Go completions are
discarded. Such an operation may still finish with side effects after cancellation.

For `FromDuplex`, ownership can close a separate writer at writable EOF while
its reader continues. Shared `FromReadWriter` endpoints may Read/Write
concurrently and must support that. On writable EOF, an owned shared endpoint's
`CloseWrite() error` is used when available. Without half-close, the shared
endpoint stays open for reads until reader EOF or explicit destroy. Both forms
use allowHalfOpen:true; callers finish/destroy both halves as appropriate. A
shared Closer is closed once rather than once per direction.

Go failures use ERR_GO_IO, short writes ERR_GO_SHORT_WRITE, and canceled contexts
ABORT_ERR. Destroying with an existing JS Error preserves that original object.
An interrupted pending write can receive ERR_STREAM_DESTROYED; queued native
callbacks settle once. Register error listeners or handle pipeline rejections.

## Example and tests

The example exports a Go MultiReader, a CopyTo SHA-256 function, and an io.Pipe
Duplex echo. It verifies binary data, promises and standard Node consumers.

```sh
make examples
node internal/examples/stream_io/index.cjs
make test-stream-io
make check
```

The integration suite covers native types, demand/backpressure, EOF with data,
empty streams, blocked reads/writes, callbacks, ownership/flush/close, mutable JS
buffers, errors/panics, deadlines, pipelines in both directions, shared
half-close, forced GC, event-loop exit and Worker termination. It uses local
memory/pipe/net.Pipe endpoints; no external connections or files are needed.

API references: [Go io](https://pkg.go.dev/io),
[Node stream implementers](https://nodejs.org/api/stream.html#api-for-stream-implementers),
[Node-API TSFN](https://nodejs.org/api/n-api.html#asynchronous-thread-safe-function-calls).
