package stream

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sync"
	"sync/atomic"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

// IOOptions configures binary Node streams backed by Go I/O. Zero sizes use
// 32 KiB. Ownership stays with the caller unless CloseOnDestroy is true. Owned
// Closers must support Close concurrently with Read/Write to interrupt blocked
// I/O. A plain io.Reader/Writer cannot be forcibly interrupted by a context.
type IOOptions struct {
	Context        context.Context
	ChunkSize      int
	HighWaterMark  int
	CloseOnDestroy bool
}

// FromReader creates a real Node Readable. Read runs in Go goroutines, only
// after Node requests data; TSFN delivers copied Buffers and EOF on Node's thread.
func (m *Module) FromReader(reader io.Reader, options ...IOOptions) (*Stream, error) {
	if nilIO(reader) {
		return nil, fmt.Errorf("FromReader requires a non-nil io.Reader")
	}
	return m.fromIO(reader, nil, false, options)
}

// FromWriter creates a real Node Writable. Node serializes writes and retains
// backpressure until the Go Write finishes. A short nil-error write fails with
// io.ErrShortWrite. A writer with Flush() error is flushed by _final.
func (m *Module) FromWriter(writer io.Writer, options ...IOOptions) (*Stream, error) {
	if nilIO(writer) {
		return nil, fmt.Errorf("FromWriter requires a non-nil io.Writer")
	}
	return m.fromIO(nil, writer, false, options)
}

// FromDuplex uses separate Go reader/writer endpoints. It keeps Node's halves
// independent (allowHalfOpen:true). With ownership enabled, _final closes the
// writer while reads continue; _destroy closes both endpoints.
func (m *Module) FromDuplex(reader io.Reader, writer io.Writer, options ...IOOptions) (*Stream, error) {
	if nilIO(reader) || nilIO(writer) {
		return nil, fmt.Errorf("FromDuplex requires a non-nil reader and writer")
	}
	return m.fromIO(reader, writer, sameIO(reader, writer), options)
}

// FromReadWriter wraps a shared endpoint. Read/Write may run concurrently. At
// writable EOF an owned CloseWrite() endpoint is half-closed when supported;
// otherwise destroy is needed to close a shared endpoint still awaiting EOF.
func (m *Module) FromReadWriter(endpoint io.ReadWriter, options ...IOOptions) (*Stream, error) {
	if nilIO(endpoint) {
		return nil, fmt.Errorf("FromReadWriter requires a non-nil io.ReadWriter")
	}
	return m.fromIO(endpoint, endpoint, true, options)
}

// CopyTo pipelines this Node readable into a Go io.Writer and returns a native
// Promise. The caller awaits it in JS or attaches Then/Catch on the Node thread.
// Node pipeline handles backpressure and destroys participants on errors.
func (s *Stream) CopyTo(writer io.Writer, options ...IOOptions) (*nodejs.Promise, error) {
	module, err := Load(s.Env())
	if err != nil {
		return nil, err
	}
	promises, err := nodejs.Load(s.Env(), "node:stream/promises")
	if err != nil {
		return nil, err
	}
	destination, err := module.FromWriter(writer, options...)
	if err != nil {
		return nil, err
	}
	return promises.CallPromise("pipeline", s, destination)
}

// CopyFrom pipelines a Go io.Reader into this Node writable asynchronously.
func (s *Stream) CopyFrom(reader io.Reader, options ...IOOptions) (*nodejs.Promise, error) {
	module, err := Load(s.Env())
	if err != nil {
		return nil, err
	}
	promises, err := nodejs.Load(s.Env(), "node:stream/promises")
	if err != nil {
		return nil, err
	}
	source, err := module.FromReader(reader, options...)
	if err != nil {
		return nil, err
	}
	return promises.CallPromise("pipeline", source, s)
}

type ioResult struct {
	id   uint64
	data []byte
	err  error
	stop bool
}

type ioOperation struct {
	kind     string
	callback *napi.Refence
	errorRef *napi.Refence
}

// Fields below the comment are confined to Node's thread. Goroutines touch only
// ctx, finalized, fn and the independent Go endpoints/ownedCloser objects.
type ioBridge struct {
	ctx        context.Context
	cancel     context.CancelFunc
	reader     io.Reader
	writer     io.Writer
	options    IOOptions
	shared     bool
	readClose  *ownedCloser
	writeClose *ownedCloser
	fn         *napi.ThreadsafeFunction
	finalized  atomic.Bool
	// Node-thread fields:
	stream     *napi.Refence // weak while idle; each operation acquires one ref
	pending    map[uint64]ioOperation
	next       uint64
	reading    bool
	readEnded  bool
	destroying bool
}

type ownedCloser struct {
	value io.Closer
	once  sync.Once
	err   error
}

func (closer *ownedCloser) close() error {
	if closer == nil {
		return nil
	}
	closer.once.Do(func() {
		closer.err = ioGuard("Close", closer.value.Close)
	})
	return closer.err
}

func (m *Module) fromIO(reader io.Reader, writer io.Writer, shared bool, options []IOOptions) (*Stream, error) {
	if len(options) > 1 {
		return nil, fmt.Errorf("at most one IOOptions value is accepted")
	}
	opt := IOOptions{}
	if len(options) == 1 {
		opt = options[0]
	}
	if opt.ChunkSize < 0 || opt.HighWaterMark < 0 {
		return nil, fmt.Errorf("I/O buffer sizes must be non-negative")
	}
	if opt.ChunkSize == 0 {
		opt.ChunkSize = 32 * 1024
	}
	if opt.HighWaterMark == 0 {
		opt.HighWaterMark = 32 * 1024
	}
	if opt.Context == nil {
		opt.Context = context.Background()
	}
	if err := opt.Context.Err(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(opt.Context)
	bridge := &ioBridge{ctx: ctx, cancel: cancel, reader: reader, writer: writer, options: opt, shared: shared, pending: make(map[uint64]ioOperation)}
	var readClose, writeClose *ownedCloser
	if opt.CloseOnDestroy {
		if closer, ok := reader.(io.Closer); ok {
			readClose = &ownedCloser{value: closer}
		}
		if closer, ok := writer.(io.Closer); ok {
			writeClose = &ownedCloser{value: closer}
			if shared && readClose != nil {
				writeClose = readClose
			}
		}
	}
	props := map[string]any{
		"objectMode": false, "decodeStrings": true, "autoDestroy": true,
		"emitClose": true, "allowHalfOpen": true, "highWaterMark": opt.HighWaterMark,
		"destroy": napi.Callback(bridge.destroyCallback),
	}
	class := "Duplex"
	if reader != nil {
		props["read"] = napi.Callback(bridge.readCallback)
	} else {
		class = "Writable"
	}
	if writer != nil {
		props["write"] = napi.Callback(bridge.writeCallback)
		props["final"] = napi.Callback(bridge.finalCallback)
	} else {
		class = "Readable"
	}
	object, err := m.New(class, props)
	if err != nil {
		cancel()
		return nil, err
	}
	bridge.stream, err = napi.CreateReference(object)
	if err != nil {
		cancel()
		return nil, err
	}
	bridge.fn, err = napi.CreateThreadsafeFunction(m.Env(), nil,
		func(env napi.EnvType, _ any) { bridge.cleanup(env) },
		func(env napi.EnvType, _ *napi.Function, data any) { bridge.deliver(env, data.(ioResult)) },
		"napi-go/stream/io", 8, 1, nil)
	if err != nil {
		cancel()
		_ = bridge.stream.RemoveRef(m.Env())
		return nil, err
	}
	if err := bridge.fn.Unref(m.Env()); err != nil {
		cancel()
		_ = bridge.fn.Release(napi.Abort)
		return nil, err
	}
	if err := object.Wrap(bridge, func(data any) {
		// Basic GC finalizers must not call Node-API or perform blocking I/O.
		state := data.(*ioBridge)
		go func() { state.cancel(); _ = state.fn.Release(napi.Abort) }()
	}); err != nil {
		cancel()
		_ = bridge.fn.Release(napi.Abort)
		return nil, err
	}
	// Ownership transfers only after the whole constructor succeeds.
	bridge.readClose, bridge.writeClose = readClose, writeClose
	if opt.Context.Done() != nil {
		go func() {
			<-ctx.Done()
			if err := opt.Context.Err(); err != nil {
				bridge.post(ioResult{stop: true, err: err})
			}
		}()
	}
	return &Stream{object}, nil
}

func (b *ioBridge) begin(env napi.EnvType, kind string, callback napi.ValueType) (uint64, error) {
	op := ioOperation{kind: kind}
	if callback != nil {
		ref, err := strongIORef(callback)
		if err != nil {
			return 0, err
		}
		op.callback = ref
	}
	if err := b.stream.Ref(env); err != nil {
		if op.callback != nil {
			_ = op.callback.RemoveRef(env)
		}
		return 0, err
	}
	if err := b.fn.Ref(env); err != nil {
		_ = b.stream.Unref(env)
		if op.callback != nil {
			_ = op.callback.RemoveRef(env)
		}
		return 0, err
	}
	b.next++
	b.pending[b.next] = op
	return b.next, nil
}

func (b *ioBridge) readCallback(ci *napi.CallbackInfo) (napi.ValueType, error) {
	if b.reading || b.readEnded || b.destroying {
		return nil, nil
	}
	id, err := b.begin(ci.Env, "read", nil)
	if err != nil {
		return nil, destroyIO(ci.Env, napi.ToObject(ci.This), err)
	}
	b.reading = true
	go func() {
		data := make([]byte, b.options.ChunkSize)
		var n int
		err := ioGuard("Read", func() error {
			for empty := 0; ; empty++ {
				if err := b.ctx.Err(); err != nil {
					return err
				}
				var err error
				n, err = b.reader.Read(data)
				if n < 0 || n > len(data) {
					n = 0
					return fmt.Errorf("io.Reader returned an invalid count")
				}
				if n > 0 || err != nil {
					return err
				}
				if empty >= 99 {
					return io.ErrNoProgress
				}
			}
		})
		b.post(ioResult{id: id, data: data[:n], err: err})
	}()
	return nil, nil
}

func (b *ioBridge) writeCallback(ci *napi.CallbackInfo) (napi.ValueType, error) {
	callback := ci.Args[2]
	data, err := nodejs.CopyBytes(ci.Args[0]) // Node decodeStrings:true yields a Buffer
	if err != nil {
		return nil, callIOCallback(ci.Env, callback, nil, err)
	}
	if b.destroying {
		return nil, callIOCallback(ci.Env, callback, nil, io.ErrClosedPipe)
	}
	id, err := b.begin(ci.Env, "write", callback)
	if err != nil {
		return nil, callIOCallback(ci.Env, callback, nil, err)
	}
	go func() {
		err := ioGuard("Write", func() error {
			if err := b.ctx.Err(); err != nil {
				return err
			}
			n, err := b.writer.Write(data)
			if n < 0 || n > len(data) {
				return fmt.Errorf("io.Writer returned an invalid count")
			}
			if n != len(data) && err == nil {
				return io.ErrShortWrite
			}
			return err
		})
		b.post(ioResult{id: id, err: err})
	}()
	return nil, nil
}

func (b *ioBridge) finalCallback(ci *napi.CallbackInfo) (napi.ValueType, error) {
	id, err := b.begin(ci.Env, "final", ci.Args[0])
	if err != nil {
		return nil, callIOCallback(ci.Env, ci.Args[0], nil, err)
	}
	go func() {
		err := ioGuard("final", func() error {
			if err := b.ctx.Err(); err != nil {
				return err
			}
			if flusher, ok := b.writer.(interface{ Flush() error }); ok {
				if err := flusher.Flush(); err != nil {
					return err
				}
			}
			if b.options.CloseOnDestroy && b.shared {
				if closer, ok := b.writer.(interface{ CloseWrite() error }); ok {
					return closer.CloseWrite()
				}
				return nil
			}
			return b.writeClose.close()
		})
		b.post(ioResult{id: id, err: err})
	}()
	return nil, nil
}

func (b *ioBridge) destroyCallback(ci *napi.CallbackInfo) (napi.ValueType, error) {
	b.destroying = true
	b.cancel()
	id, err := b.begin(ci.Env, "destroy", ci.Args[1])
	if err != nil {
		return nil, callIOCallback(ci.Env, ci.Args[1], ci.Args[0], err)
	}
	kind, err := ci.Args[0].Type()
	if err == nil && kind != napi.TypeNull && kind != napi.TypeUndefined {
		var ref *napi.Refence
		ref, err = strongIORef(ci.Args[0])
		if err == nil {
			op := b.pending[id]
			op.errorRef = ref
			b.pending[id] = op
		}
	}
	if err != nil {
		go b.post(ioResult{id: id, err: err})
		return nil, nil
	}
	go func() { b.post(ioResult{id: id, err: b.closeOwned()}) }()
	return nil, nil
}

func (b *ioBridge) post(result ioResult) {
	if !b.finalized.Load() {
		_ = b.fn.Call(result, napi.Blocking)
	}
}

func (b *ioBridge) deliver(env napi.EnvType, result ioResult) {
	if result.stop {
		if !b.destroying {
			if value, err := b.stream.ValeuOf(env); err == nil && value.NapiValue() != nil {
				_ = destroyIO(env, napi.ToObject(value), result.err)
			}
		}
		return
	}
	op, ok := b.pending[result.id]
	if !ok {
		return
	}
	delete(b.pending, result.id)
	if op.kind == "read" {
		b.reading = false
		b.readEnded = result.err != nil
		if !b.destroying {
			value, err := b.stream.ValeuOf(env)
			if err == nil && value.NapiValue() != nil {
				object := napi.ToObject(value)
				if len(result.data) > 0 {
					buffer, copyErr := napi.CopyBuffer(env, result.data)
					if copyErr == nil {
						_, copyErr = object.CallMethod("push", buffer)
					}
					if copyErr != nil {
						result.err = copyErr
					}
				}
				if result.err == io.EOF {
					_, _ = object.CallMethod("push", mustIONull(env))
				} else if result.err != nil {
					_ = destroyIO(env, object, result.err)
				}
			}
		}
	} else {
		if op.kind == "destroy" {
			// Settle outstanding native write/final callbacks even if their Go
			// endpoint is uninterruptible. Late Go completions are then ignored.
			for id, abandoned := range b.pending {
				delete(b.pending, id)
				if abandoned.callback != nil {
					callback, err := abandoned.callback.ValeuOf(env)
					if err == nil && callback.NapiValue() != nil {
						_ = callIOCallback(env, callback, nil, io.ErrClosedPipe)
					}
					_ = abandoned.callback.RemoveRef(env)
				}
				if abandoned.errorRef != nil {
					_ = abandoned.errorRef.RemoveRef(env)
				}
				_ = b.stream.Unref(env)
			}
		}
		var original napi.ValueType
		if op.errorRef != nil {
			original, _ = op.errorRef.ValeuOf(env)
		}
		callback, err := op.callback.ValeuOf(env)
		if err == nil && callback.NapiValue() != nil {
			if b.destroying && op.kind != "destroy" {
				result.err = io.ErrClosedPipe
			}
			_ = callIOCallback(env, callback, original, result.err)
		}
	}
	if op.callback != nil {
		_ = op.callback.RemoveRef(env)
	}
	if op.errorRef != nil {
		_ = op.errorRef.RemoveRef(env)
	}
	_ = b.stream.Unref(env)
	if op.kind == "destroy" {
		_ = b.fn.Release(napi.Abort)
	} else if len(b.pending) == 0 {
		_ = b.fn.Unref(env)
	}
}

func (b *ioBridge) cleanup(env napi.EnvType) {
	b.finalized.Store(true)
	b.cancel()
	if env != nil {
		for _, op := range b.pending {
			if op.callback != nil {
				_ = op.callback.RemoveRef(env)
			}
			if op.errorRef != nil {
				_ = op.errorRef.RemoveRef(env)
			}
		}
		_ = b.stream.RemoveRef(env)
	}
	b.pending = nil
	go func() { _ = b.closeOwned() }()
}

func (b *ioBridge) closeOwned() error {
	if b.readClose == b.writeClose {
		return b.readClose.close()
	}
	// Close endpoints concurrently: either can be needed to wake the other side.
	read := make(chan error, 1)
	go func() { read <- b.readClose.close() }()
	writeErr := b.writeClose.close()
	return errors.Join(<-read, writeErr)
}

func strongIORef(value napi.ValueType) (*napi.Refence, error) {
	ref, err := napi.CreateReference(value)
	if err != nil {
		return nil, err
	}
	if err := ref.Ref(value.Env()); err != nil {
		_ = ref.RemoveRef(value.Env())
		return nil, err
	}
	return ref, nil
}

func callIOCallback(env napi.EnvType, callback, original napi.ValueType, err error) error {
	if original != nil && original.NapiValue() != nil {
		kind, typeErr := original.Type()
		if typeErr != nil {
			return typeErr
		}
		if kind == napi.TypeNull || kind == napi.TypeUndefined {
			original = nil
		}
	}
	if original == nil || original.NapiValue() == nil {
		var conversion error
		original, conversion = ioErrorValue(env, err)
		if conversion != nil {
			return conversion
		}
	}
	_, callErr := napi.ToFunction(callback).Call(original)
	return callErr
}

func ioErrorValue(env napi.EnvType, err error) (napi.ValueType, error) {
	if err == nil {
		return env.Null()
	}
	value, createErr := napi.CreateError(env, err.Error())
	if createErr != nil {
		return nil, createErr
	}
	code := "ERR_GO_IO"
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		code = "ABORT_ERR"
	}
	if errors.Is(err, io.ErrShortWrite) {
		code = "ERR_GO_SHORT_WRITE"
	}
	if errors.Is(err, io.ErrClosedPipe) {
		code = "ERR_STREAM_DESTROYED"
	}
	if setErr := napi.ToObject(value).SetValue("code", code); setErr != nil {
		return nil, setErr
	}
	return value, nil
}

func destroyIO(env napi.EnvType, object *napi.Object, err error) error {
	value, conversion := ioErrorValue(env, err)
	if conversion != nil {
		return conversion
	}
	_, callErr := object.CallMethod("destroy", value)
	return callErr
}

func mustIONull(env napi.EnvType) napi.ValueType {
	value, err := env.Null()
	if err != nil {
		panic(err)
	}
	return value
}

func ioGuard(operation string, callback func() error) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			err = fmt.Errorf("Go I/O %s panic: %v", operation, recovered)
		}
	}()
	return callback()
}

func nilIO(value any) bool {
	if value == nil {
		return true
	}
	ptr := reflect.ValueOf(value)
	switch ptr.Kind() {
	case reflect.Pointer, reflect.Map, reflect.Slice, reflect.Func, reflect.Interface, reflect.Chan:
		return ptr.IsNil()
	}
	return false
}

func sameIO(reader io.Reader, writer io.Writer) bool {
	typeOf := reflect.TypeOf(reader)
	return typeOf == reflect.TypeOf(writer) && typeOf.Comparable() && any(reader) == any(writer)
}
