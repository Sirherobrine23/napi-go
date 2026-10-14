# Node.js wrappers validation

Apply this incremental patch after `napi-go-object-helpers.patch`. Its base is
upstream `6778306b09b8069e7b20c37ec4a81f94099a1924` plus that object-helper patch
(local tree `520e7cde9f67782978747394566982b7925df044`).

New typed packages: path, os, fs, buffer, crypto, url, util, zlib, timers, stream,
dns. Existing EventEmitter gains Go-value/function helpers. A common module API
exposes other builtins and exports without needing a separate typed wrapper.
The JS-owned Promise wrapper supports Then/Catch/Finally and adoption without
treating Promise handles as native deferred resources.

Validation environment: Linux amd64, Go 1.26.8, Node.js 24.19.0, matching Node
headers and a C compiler. Commands:

```sh
make check examples
node internal/examples/nodejs/index.cjs
make test-nodejs
```

- All Go packages and seven example addons build.
- 84 Node integration tests pass: the existing 60 plus 24 builtin tests.
- All 84 tests pass under `GOEXPERIMENT=cgocheck2` with no skipped tests.
- `go vet ./nodejs/...` passes; the preexisting raw-finalizer warnings outside
  these packages keep repository-wide vet separate from make check.
- The new example successfully hashes its own file through fs/promises and a
  Go Promise callback, verifies it against Node crypto, and checks path/timers.
- Changed Go files are formatted and `git diff --check` passes.

Covered behaviors include actual builtin identity, generic dispatch, receiver
preservation, POSIX/Windows paths, CPU/user decoding, network-interface results
or original host permission errors, Buffer encodings/copies and empty data,
hashing/HMAC/random bytes, WHATWG URLs/search parameters, promisify/callbackify,
gzip/deflate/Brotli round-trips, sync/async filesystem operations and error codes,
optional missing stat results, Promise adoption/failure/finally, timer callbacks
and clearing/ref/unref, stream backpressure/pipe/listeners, localhost DNS, ESM,
Workers and the legacy explicit global-require loader.

The host denies os.networkInterfaces with a native uv_interface_addresses
permission error. The wrapper test verifies the same code/syscall as direct
Node; it does not skip the operation or turn the error into an empty result.
Public DNS/network connections are not used by the suite. Every test filesystem
mutation takes place in a freshly created temporary directory cleaned by Node.

Required binary binding fixes are included: zero-length CreateBufferCopy passes
a nil source pointer, GetBufferInfo/Data writes the native pointer into the actual
Go output variable, and Type reports Buffer before TypedArray. CopyBytes returns
independent Go-owned memory; Buffer.Data remains a view of JS memory.

Local validation used NODE_INCLUDE for the parent of node/node_api.h and
GOFLAGS=-buildvcs=false for unavailable VCS stamping in the detached worktree.
Native cgo checking and linker behavior were preserved. Race-runtime testing was
not rerun; the earlier environment limitation still applies.

All wrappers require the owning Node thread and current callback scope. They
are not goroutine-safe and introduce no synchronous Go wait for JS completion.
The sync fs/zlib/crypto operations still block the Node thread; use native
Promises/callbacks or a Go TSFN/EventEmitterThread producer as appropriate.

The typed surface covers common operations, not every Node export. Generic
Module.Call/New/Get remains available for additional exports, option forms and
version-specific functions. The existing generic require.Import implementation
is not used or changed by these wrappers.
