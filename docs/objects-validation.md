# Object helpers validation

Incremental base: `6778306b09b8069e7b20c37ec4a81f94099a1924` (`Add event threads`).
Existing object, class, function and thread-safe APIs remain available.

The patch adds Go-value helpers for construction, named properties, ordinary Go
methods, receiver-preserving calls, typed argument/property/result decoding and
checked wrapped state. It also repairs the conversions these helpers use:
slice completion, nested `any` maps, pointer allocation, named scalar types,
unnamed native callbacks and fixed/variadic argument validation. Conversion
recursion is bounded and nil/foreign handles are validated before native calls.

Validated on Linux amd64 with Go 1.26.8, Node.js 24.19.0 and matching Node headers:

```sh
make check examples
node internal/examples/objects/index.cjs
```

- All Go packages build.
- All 60 Node integration tests pass: the existing 45 plus 15 object tests.
- The same 60 tests pass with `GOEXPERIMENT=cgocheck2`.
- All six example addons build. The new objects example verifies increments,
  typed struct input and an invalid argument; the updated Counter example works.
- `git diff --check` passes and changed Go files are formatted.

The object suite covers nested Go maps/structs/slices, tagged and pointer fields,
typed nil handles, native and ordinary Go functions, variadic and fixed arity,
Go errors, checked wrappers, aliases, invalid decode targets, cyclic data in both
directions, foreign environments, method receivers, inherited methods, zero-arg
calls, JavaScript exceptions, class construction and wrapped state mismatches.
Assignment tests include deterministic setter order, conversion/setter failures
after earlier assignments, and the existing behavior on frozen objects.

Tests exercise a real c-shared addon through Node. In this exported worktree,
`GOFLAGS=-buildvcs=false` disables unavailable VCS stamping and `NODE_INCLUDE`
points at the parent of `node/node_api.h`. No linker or cgo validation was disabled.
Race-runtime validation was not rerun for this patch. The existing three raw
finalizer warnings still make `make vet` a separate target.

Object handles remain bound to their Node environment and scope. For background
work use `EventEmitterThread`/TSFN with Go-owned payloads; call these helpers only
after delivery on the Node thread. Conversion and assignment are not transactional.
Numeric conversions keep the existing Number/BigInt and truncation rules.
