# Class and ThreadsafeFunction implementation

Base: `f5b390708d7c95ffb71034ba7873cd70b892ee28` (`main`).
Validated on Linux amd64 with Go 1.26.8, GCC and Node.js v24.19.0,
using the official matching Node development headers.

## Implemented

- `CreateClass`, `Class.New`, `ToClass` and `As[*Class]`.
- Constructor/method/accessor callbacks share the existing `CallbackInfo` API.
- Static/prototype properties, property attributes and symbol property names.
- `Object.DefineProperties`, `Wrap`, `Unwrap` and `RemoveWrap`.
- Go instance ownership through C-allocated tokens and `runtime/cgo.Handle`.
- Thread-safe queues accept existing JavaScript functions or Go callbacks.
- Blocking/nonblocking calls, Go data delivery, queue-full/closing errors,
  producer acquisitions, release/abort, context and event-loop ref/unref.
- Finalization retires remaining producer acquisitions and waits for native
  calls already in progress before letting Node destroy the native function.
- Cleanup of rejected/abandoned payloads and cleanup without a user finalizer.
- Independent Worker environments and forced termination with active producers.

A bounded queue exposed an intermittent stall with several native blocking
waiters. Blocking calls now serialize access to the native wait; other producers
wait in Go. Abort does not take that mutex and wakes the native waiter.

## Checks

Passed:

```sh
go build ./...
go test ./...
GOEXPERIMENT=cgocheck2 go test -count=1 -v ./internal/integration
git diff --check
```

The integration runner builds a real `.node` addon and runs 17 JavaScript tests.
It checks constructors/subclasses, native construction with no arguments,
properties, wrapped state GC/removal, extracted callbacks, concurrent producers,
queue-full behavior, abort (including four blocked producers), finalization,
payload conversion, context, event-loop liveness and Worker teardown. One test
checks reclamation of 1,000 Go payloads discarded by Abort.

`go vet ./...` still reports three existing issues: Go function values cast to C
finalizer pointers in the old internal `CreateExternal`,
`CreateExternalArrayBuffer` and `Wrap` APIs. These exact warnings were also found
on the unmodified base (which additionally had three unsafe integer/pointer
conversions removed by this change). The new public Go-state wrap API uses a C
trampoline instead of that old raw-pointer internal wrapper.

An addon built with `-race` compiled successfully, but ThreadSanitizer could not
allocate its shadow memory in this execution environment. The race-instrumented
suite could not run; this is not claimed as a passing check.

## Compatibility and scope

`ThreadsafeFunction` no longer implements `ValueType`: it is a native queue,
not a JavaScript value. Export an entry function that creates the queue and
starts the producers. The existing example has been updated accordingly.
`CreateThreadsafeFunction` keeps the parameter order and accepts the old Go
`Callback` form as well as JavaScript function values.

Class conversion is explicit. Reflection-based struct conversion still creates
plain objects; automatic Go-method discovery is not added. Class/descriptor
callbacks live until their environment exits, which keeps extracted methods and
accessors valid. Wrapped instance data is collected independently.

All JavaScript values retain Node-API's environment/thread/handle-scope lifetime.
Do not retain raw `Class`/`Function` handles across callbacks without a native
reference, or send JavaScript handles (including in containers) as queue data.
