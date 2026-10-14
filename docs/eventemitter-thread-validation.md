# EventEmitterThread validation

This incremental applies after the EventEmitter tests/Makefile incremental and
`napi-go-status-error-compat.patch`, on top of upstream
`8c4ad95f0037876ac6d9645b28532c3f4b9d7a29`.

The native queue provides nonblocking and blocking event emission, automatic
processor goroutines, context cancellation, drain/abort, native completion and
event-loop references. Dispatch uses a JavaScript `emit` bound to its emitter,
with a native reference for scoped access to that object. Go payload conversion
and JavaScript invocation happen on the Node thread.

The existing synchronous EventEmitter constructor and method receivers were
corrected. `Emit` now includes the event name. A synchronous builtin loader uses
`process.getBuiltinModule` or an explicit global `require`. Three internal
reference operations now pass `ref.Ref` instead of the address of the Go struct;
these fixes are necessary to keep an emitter alive between callback scopes.

Validated with Go 1.26.8, Node.js 24.19.0, matching official headers and GCC on
Linux amd64. All 45 integration cases pass, including strict cgo validation:

```sh
make check examples
make test-eventemitter-thread
node internal/examples/events_thread/index.cjs
node -e 'const a=require("./build/events.node"); const {EventEmitter}=require("node:events"); if (!(a.events instanceof EventEmitter)) throw Error("wrong emitter");'
git diff --check
```

Explicit GO and NODE_INCLUDE overrides were used in this environment. Its
exported/worktree source metadata required GOFLAGS=-buildvcs=false; the runtime
checks and example addon builds still execute normally. Previous limitations
of race instrumentation and existing raw-finalizer vet warnings still apply.

The 19 new native-emitter cases cover:

- Native construction without global require, event names/arguments, listener
  receivers, once/prepend ordering, removal and emitter query methods.
- Background processors and concurrent Go producers, plus graceful queue drain.
- Nonblocking queue saturation and rejection of direct/nested JavaScript handles
  and cyclic Go payloads before leaving the Node callback.
- Abort, context cancellation, processor errors, recovered Go panics and manual
  Go errors converted into JavaScript Error values.
- Fresh emitter handles after callback scope exit and forced GC.
- Repeated/invalid processors, idempotent close/abort and calls after closure.
- Unreferenced queues at process exit and Worker termination during processing.
- ESM, explicit global-require fallback and a clear missing-loader error.

The processor must respect cancellation and join its child producers before it
returns. Neither Go goroutines nor JavaScript event listeners run in a separate
JavaScript isolate; listeners execute on the emitter's owning Node thread.
