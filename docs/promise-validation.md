# Promise chaining regression

Incremental patch base: `36e40e89beccf333e69df2850bb74642e8cba586`.

The new guard in `nodejs.Promise.Then` rejected `Then(nil, nil)` with
`add one of functions`. This broke the existing builtin integration test and
also `Catch(nil)`, whose fallback called `Then(nil, nil)`.

Literal nil handlers now become JavaScript `undefined` for Then/Catch/Finally.
Native Promise methods create the new chain and forward the settlement, as
specified by [ECMAScript](https://tc39.es/ecma262/2025/multipage/control-abstraction-objects.html#sec-promise.prototype.then).
Non-nil handlers still require Go or JavaScript functions.

Catch invokes the object's catch method directly, including when its handler is
nil. This preserves overrides and lookup exceptions, avoids extra property
existence traps, and reports missing methods instead of silently falling back.

Validation: Linux amd64, Go 1.26.8, Node.js 24.19.0 with matching Node headers.

```sh
make test-nodejs
make check examples
go vet ./nodejs/...
```

- The original code failed the existing Promise integration test.
- Seven added tests cover distinct chains, value/rejection identity, optional
  callbacks, asynchronous reactions, thenable adoption, cleanup waiting and
  failure, catch overrides, property traps, getter exceptions and invalid inputs.
- All 116 integration tests pass normally and with `GOEXPERIMENT=cgocheck2`.
- All eight example addons build; Node wrapper vet and `git diff --check` pass.

Local builds supplied `NODE_INCLUDE` for the matching headers and
`GOFLAGS=-buildvcs=false` for the detached worktree's VCS stamping.
