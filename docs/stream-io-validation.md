# Go io / Node stream bridge validation

This incremental patch applies after `napi-go-nodejs-modules.patch`. Prepared
against upstream `6a99dd5` plus that module patch (local base tree
`0adb31902591f9de5fc19ceb9c5e345a4a24cbe4`). The earlier Node module patch applies
cleanly to that upstream revision.

The bridge adds Go-backed Readable/Writable/Duplex constructors and Promise
CopyTo/CopyFrom pipelines. Existing Stream Read/Write signatures remain intact.
I/O and owned resource closing run in Go goroutines; only Node callbacks resolve
references, call stream methods and invoke native write/final/destroy callbacks.

Validated on Linux amd64 with Go 1.26.8, Node.js 24.19.0 and matching Node headers:

```sh
make check examples
make test-stream-io
node internal/examples/stream_io/index.cjs
go vet ./nodejs/stream
```

- All Go packages and eight example addons build.
- All 109 Node integration tests pass, including the 25 new Go io tests.
- The same 109 tests pass with GOEXPERIMENT=cgocheck2, with no skipped tests.
- The stream package passes go vet.
- The example verifies MultiReader -> Node, Node readable -> Go SHA-256 writer,
  and binary echo over an io.Pipe Duplex.
- Changed Go files are formatted; git diff --check passes.

The tests cover demand-only reads, bounded prefetch, ordered chunks, empty input,
bytes with EOF, reader errors/panics/invalid counts/no progress, event-loop
responsiveness during slow Go I/O, mutable JS buffer copies, serialized writes,
native backpressure/acknowledgement, write/flush/close errors, ownership and
exactly-once closing, context cancellation/deadlines, pending/buffered callback
settlement, separate pipes/shared net.Pipe endpoints, CloseWrite half-close,
pipeline in both directions, idle event-loop exit, weak-reference GC cleanup
and Worker termination during an owned blocked read.

References to stream objects are weak while idle and acquired for active
operations. Callback/error references are strong until their native callback is
settled or the Node environment is torn down. A bounded TSFN delivers only Go
bytes/error/id values, never scoped JS handles. GC cleanup performs no Node-API
or blocking I/O in a basic wrap finalizer. Late I/O results are dropped after
closure, including when an uninterruptible borrowed writer later returns.

No test uses remote network services or filesystem resources. net.Pipe and
io.Pipe are in-memory endpoints. Local validation sets NODE_INCLUDE to the
parent of node/node_api.h and GOFLAGS=-buildvcs=false for unavailable detached
worktree VCS stamping; no native linker/cgo validation was disabled.

Race-runtime testing was not repeated: the earlier native ThreadSanitizer
allocation limitation in this environment still applies. Repository-wide vet
remains separate because of the preexisting raw-finalizer warnings elsewhere.

Cancellation cannot forcibly stop an arbitrary io.Reader/Writer. Ownership is
explicit through CloseOnDestroy; owned Closers must support Close concurrently
with I/O and unblock it. Shared read/write endpoints must support concurrent
Read/Write. Native completion can settle before an uninterruptible borrowed
operation returns; its caller is responsible for unblocking that operation.
Native helpers still require the owning Node thread and live callback scope.
