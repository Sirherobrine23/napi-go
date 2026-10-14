package napi

import (
	"fmt"
	"reflect"
	"sync"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// ThreadsafeFunction is a native queue, not a JavaScript ValueType. Create it on
// the Node thread and use Call from goroutines. Each producer must own an initial
// acquisition or Acquire one before starting, and Release it when finished.
// Do not pass ValueTypes as data: JavaScript handles belong to their Node scope.
type ThreadsafeFunction struct {
	mu         sync.Mutex
	blockingMu sync.Mutex
	idle       *sync.Cond
	inflight   int
	env        EnvType
	tsfn       napi.ThreadsafeFunction
	users      int
	closing    bool
	context    any
}

type ThreadsafeFunctionReleaseMode = napi.ThreadsafeFunctionReleaseMode
type ThreadsafeFunctionCallMode = napi.ThreadsafeFunctionCallMode

const (
	Release     ThreadsafeFunctionReleaseMode = napi.Release
	Abort       ThreadsafeFunctionReleaseMode = napi.Abort
	NonBlocking ThreadsafeFunctionCallMode    = napi.NonBlocking
	Blocking    ThreadsafeFunctionCallMode    = napi.Blocking
)

// These errors support errors.Is without importing internal/napi.
var (
	ErrThreadsafeFunctionClosing   error = napi.StatusError(napi.StatusClosing)
	ErrThreadsafeFunctionQueueFull error = napi.StatusError(napi.StatusQueueFull)
)

// A finalizer runs once after the queue is drained. env may be nil at teardown;
// do not execute JavaScript then. Internal cleanup runs even without a finalizer.
type ThreadsafeFunctionFinalizeCallback func(env EnvType, context any)

// The callback runs on the Node thread. jsCallback is nil when none was supplied.
// Go data can be converted to JavaScript here. It is never invoked with a nil env.
type ThreadsafeFunctionCallJSCallback func(env EnvType, jsCallback *Function, data any)

// CreateThreadsafeFunction accepts an existing JavaScript Function/ValueType, a
// Callback (including an unnamed func with that signature), or nil. When
// callJSCallback is nil, jsFunc receives the converted data as one argument;
// nil data produces a zero-argument call. At least one callback is required.
// maxQueueSize == 0 means unbounded. initialThreadCount must be >= 1.
func CreateThreadsafeFunction(env EnvType, jsFunc any, finalizeCallback ThreadsafeFunctionFinalizeCallback, callJSCallback ThreadsafeFunctionCallJSCallback, resourceName string, maxQueueSize, initialThreadCount int, context any) (*ThreadsafeFunction, error) {
	if maxQueueSize < 0 || initialThreadCount < 1 {
		return nil, fmt.Errorf("maxQueueSize must be non-negative and initialThreadCount must be at least 1")
	}
	if value := reflect.ValueOf(jsFunc); value.IsValid() && value.Kind() == reflect.Pointer && value.IsNil() {
		jsFunc = nil
	}
	var function ValueType
	var err error
	switch callback := jsFunc.(type) {
	case nil:
	case ValueType:
		function = callback
	case Callback:
		if callback != nil {
			function, err = CreateFunction(env, resourceName, callback)
		}
	case func(*CallbackInfo) (ValueType, error):
		if callback != nil {
			function, err = CreateFunction(env, resourceName, Callback(callback))
		}
	default:
		return nil, fmt.Errorf("jsFunc must be a JavaScript function, Callback or nil")
	}
	if err != nil {
		return nil, err
	}
	var functionValue napi.Value
	if function != nil {
		if function.NapiEnv() != env.NapiValue() {
			return nil, fmt.Errorf("jsFunc belongs to another environment")
		}
		typeOf, err := function.Type()
		if err != nil {
			return nil, err
		}
		if typeOf != TypeFunction {
			return nil, fmt.Errorf("jsFunc must be a JavaScript function")
		}
		functionValue = function.NapiValue()
	}
	if functionValue == nil && callJSCallback == nil {
		return nil, fmt.Errorf("jsFunc or callJSCallback must be supplied")
	}
	resource, err := CreateString(env, resourceName)
	if err != nil {
		return nil, err
	}
	if callJSCallback == nil {
		callJSCallback = func(env EnvType, callback *Function, data any) {
			var args []ValueType
			if data != nil {
				value, err := ValueOf(env, data)
				if err != nil {
					panic(err)
				}
				args = append(args, value)
			}
			if _, err := callback.Call(args...); err != nil {
				panic(err)
			}
		}
	}
	wrapper := &ThreadsafeFunction{env: env, users: initialThreadCount, context: context}
	wrapper.idle = sync.NewCond(&wrapper.mu)
	functionHandle, status := napi.CreateThreadsafeFunctionWithCallbacks(env.NapiValue(), functionValue, resource.NapiValue(), maxQueueSize, initialThreadCount,
		func(env napi.Env, callback napi.Value, data any) {
			var function *Function
			if callback != nil {
				function = ToFunction(N_APIValue(N_APIEnv(env), callback))
			}
			callJSCallback(N_APIEnv(env), function, data)
		},
		func(env napi.Env) {
			wrapper.mu.Lock()
			wrapper.closing = true
			// Abort and environment teardown may finalize before producers have
			// released all initial acquisitions. Retire them here while the
			// native object is still valid; later Go Release calls see closing.
			for i := 0; i < wrapper.users; i++ {
				napi.ReleaseThreadsafeFunction(wrapper.tsfn, Release)
			}
			wrapper.users = 0
			// Wait only for calls already inside N-API. They can finish without
			// running JS, and their temporary native acquisitions must be gone
			// before the native finalizer returns (also on older Node versions).
			for wrapper.inflight > 0 {
				wrapper.idle.Wait()
			}
			wrapper.tsfn = nil
			context := wrapper.context
			wrapper.context = nil
			wrapper.mu.Unlock()
			if finalizeCallback != nil {
				var environment EnvType
				if env != nil {
					environment = N_APIEnv(env)
				}
				finalizeCallback(environment, context)
			}
		})
	if err := status.ToError(); err != nil {
		return nil, err
	}
	wrapper.tsfn = functionHandle
	return wrapper, nil
}

// GetContext is safe from any goroutine, and fails after native finalization.
func (fn *ThreadsafeFunction) GetContext() (any, error) {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	if fn.tsfn == nil {
		return nil, ErrThreadsafeFunctionClosing
	}
	return fn.context, nil
}

// Call queues Go-owned data. Blocking must only be used from background
// goroutines: using it on the Node thread can deadlock a bounded queue.
func (fn *ThreadsafeFunction) Call(data any, mode ThreadsafeFunctionCallMode) error {
	if mode != Blocking && mode != NonBlocking {
		return fmt.Errorf("invalid threadsafe call mode")
	}
	if _, ok := data.(ValueType); ok {
		return fmt.Errorf("threadsafe data must be Go-owned, not a JavaScript handle")
	}
	if mode == Blocking {
		// Node signals one native waiter when a bounded queue transitions from
		// full. Keep one waiter inside N-API; other goroutines wait here so a
		// fast queue drain cannot strand native waiters after the last item.
		// Abort never takes this mutex and can always wake the native waiter.
		fn.blockingMu.Lock()
		defer fn.blockingMu.Unlock()
	}
	fn.mu.Lock()
	if fn.tsfn == nil || fn.closing || fn.users == 0 {
		fn.mu.Unlock()
		return ErrThreadsafeFunctionClosing
	}
	handle := fn.tsfn
	// A temporary native acquisition keeps the function alive during a blocking
	// call without holding mu. Abort on the Node thread must be able to wake a
	// producer blocked on a full queue.
	status := napi.AcquireThreadsafeFunction(handle)
	if status == napi.StatusClosing {
		fn.closing = true
	}
	if status == napi.StatusOK {
		fn.inflight++
	}
	fn.mu.Unlock()
	if err := status.ToError(); err != nil {
		return err
	}
	status = napi.CallThreadsafeFunctionWithData(handle, data, mode)
	if status == napi.StatusClosing {
		fn.mu.Lock()
		fn.closing = true
		fn.mu.Unlock()
	}
	if status != napi.StatusClosing {
		// napi_closing consumes this call's native acquisition. Never use the
		// handle again after that status, including releasing it twice.
		napi.ReleaseThreadsafeFunction(handle, Release)
	}
	fn.mu.Lock()
	fn.inflight--
	if fn.inflight == 0 {
		fn.idle.Broadcast()
	}
	fn.mu.Unlock()
	if err := status.ToError(); err != nil {
		return fmt.Errorf("call threadsafe function: %w", err)
	}
	return nil
}

func (fn *ThreadsafeFunction) Acquire() error {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	if fn.tsfn == nil || fn.closing || fn.users == 0 {
		return ErrThreadsafeFunctionClosing
	}
	if err := napi.AcquireThreadsafeFunction(fn.tsfn).ToError(); err != nil {
		return err
	}
	fn.users++
	return nil
}

// Release ends one producer's acquisition. Abort also stops new calls. Other
// producers must stop and release their own acquisitions after Abort.
func (fn *ThreadsafeFunction) Release(mode ThreadsafeFunctionReleaseMode) error {
	if mode != Release && mode != Abort {
		return fmt.Errorf("invalid threadsafe release mode")
	}
	fn.mu.Lock()
	defer fn.mu.Unlock()
	if fn.tsfn == nil || fn.users == 0 {
		return ErrThreadsafeFunctionClosing
	}
	if err := napi.ReleaseThreadsafeFunction(fn.tsfn, mode).ToError(); err != nil {
		return err
	}
	fn.users--
	if mode == Abort || fn.users == 0 {
		fn.closing = true
	}
	return nil
}

// Ref/Unref control whether the Node event loop stays alive; they do not change
// producer acquisitions. They are idempotent and must run on the Node thread.
func (fn *ThreadsafeFunction) Ref(env EnvType) error   { return fn.setRef(env, true) }
func (fn *ThreadsafeFunction) Unref(env EnvType) error { return fn.setRef(env, false) }

func (fn *ThreadsafeFunction) setRef(env EnvType, ref bool) error {
	fn.mu.Lock()
	defer fn.mu.Unlock()
	if fn.tsfn == nil || fn.closing {
		return ErrThreadsafeFunctionClosing
	}
	if env.NapiValue() != fn.env.NapiValue() {
		return fmt.Errorf("threadsafe function belongs to another environment")
	}
	if ref {
		return napi.RefThreadsafeFunction(env.NapiValue(), fn.tsfn).ToError()
	}
	return napi.UnrefThreadsafeFunction(env.NapiValue(), fn.tsfn).ToError()
}
