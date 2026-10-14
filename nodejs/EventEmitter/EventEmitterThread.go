package event

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

// EventEmitterThread is a native queue for emitting events from Go goroutines.
// It deliberately does not implement ValueType. Emitter returns a scoped JS
// value on the Node thread; Emit/EmitBlocking/Close/Abort/Run are safe from Go.
type EventEmitterThread struct {
	mu      sync.Mutex
	env     napi.EnvType
	ref     *napi.Refence
	fn      *napi.ThreadsafeFunction
	ctx     context.Context
	cancel  context.CancelFunc
	done    chan struct{}
	closing bool
	running bool
}

type threadEvent struct {
	name   string
	values []any
	err    error
}

// CreateEventThread creates a Node EventEmitter and its queue on the Node thread.
// Queue size 0 is unbounded. Close or Abort when done, or use Run for auto-close.
func CreateEventThread(env napi.EnvType, maxQueueSize int) (*EventEmitterThread, error) {
	if maxQueueSize < 0 {
		return nil, fmt.Errorf("maxQueueSize must be non-negative")
	}
	emitter, err := CreateEvent(env)
	if err != nil {
		return nil, err
	}
	return emitter.Thread(maxQueueSize)
}

// Thread attaches a queue to an existing emitter in the current Node scope.
// It retains the emitter until native finalization, even after the scope ends.
func (eve *EventEmitter) Thread(maxQueueSize int) (*EventEmitterThread, error) {
	if maxQueueSize < 0 {
		return nil, fmt.Errorf("maxQueueSize must be non-negative")
	}
	emit, err := eve.valueOf.Get("emit")
	if err != nil {
		return nil, err
	}
	bind, err := napi.ToObject(emit).Get("bind")
	if err != nil {
		return nil, err
	}
	bound, err := napi.ToFunction(bind).CallWithGlobal(emit, eve)
	if err != nil {
		return nil, err
	}
	ref, err := napi.CreateReference(eve)
	if err != nil {
		return nil, err
	}
	if err := ref.Ref(eve.env); err != nil {
		ref.RemoveRef(eve.env)
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	thread := &EventEmitterThread{env: eve.env, ref: ref, ctx: ctx, cancel: cancel, done: make(chan struct{})}
	fn, err := napi.CreateThreadsafeFunction(eve.env, bound,
		func(env napi.EnvType, _ any) {
			thread.mu.Lock()
			thread.closing = true
			thread.cancel()
			ref := thread.ref
			thread.ref = nil
			thread.mu.Unlock()
			defer close(thread.done)
			if env != nil {
				if err := ref.RemoveRef(env); err != nil {
					panic(err)
				}
			}
		},
		func(env napi.EnvType, callback *napi.Function, data any) {
			message := data.(threadEvent)
			name, err := napi.CreateString(env, message.name)
			if err != nil {
				panic(err)
			}
			args := []napi.ValueType{name}
			if message.err != nil {
				value, err := napi.CreateError(env, message.err.Error())
				if err != nil {
					panic(err)
				}
				args = append(args, value)
			} else {
				for _, arg := range message.values {
					var value napi.ValueType
					var err error
					if goError, ok := arg.(error); ok {
						value, err = napi.CreateError(env, goError.Error())
					} else {
						value, err = napi.ValueOf(env, arg)
					}
					if err != nil {
						panic(err)
					}
					args = append(args, value)
				}
			}
			if _, err := callback.Call(args...); err != nil {
				panic(err)
			}
		}, "napi-go/EventEmitterThread", maxQueueSize, 1, nil)
	if err != nil {
		cancel()
		ref.RemoveRef(eve.env)
		return nil, err
	}
	thread.fn = fn
	return thread, nil
}

// Emitter retrieves a fresh handle in the supplied Node scope. Call only from
// the owning Node thread. After finalization it returns the closing error;
// a JavaScript object already exported to JS remains valid independently.
func (thread *EventEmitterThread) Emitter(env napi.EnvType) (*EventEmitter, error) {
	thread.mu.Lock()
	defer thread.mu.Unlock()
	if thread.ref == nil {
		return nil, napi.ErrThreadsafeFunctionClosing
	}
	if env == nil || env.NapiValue() != thread.env.NapiValue() {
		return nil, fmt.Errorf("emitter belongs to another Node environment")
	}
	value, err := thread.ref.ValeuOf(env)
	if err != nil {
		return nil, err
	}
	return ToEventEmitter(value), nil
}

// Emit queues an event without waiting for queue space. The return value is
// queue acceptance, not emitter.emit's boolean. Keep Go payloads immutable
// until delivery. Node ValueTypes must not be passed, even inside containers.
func (thread *EventEmitterThread) Emit(name string, values ...any) error {
	return thread.emit(name, values, napi.NonBlocking)
}

// EmitBlocking waits for bounded queue space; use only from Go goroutines.
// Calling it from the Node thread can prevent the event loop from draining.
func (thread *EventEmitterThread) EmitBlocking(name string, values ...any) error {
	return thread.emit(name, values, napi.Blocking)
}

func (thread *EventEmitterThread) emit(name string, values []any, mode napi.ThreadsafeFunctionCallMode) error {
	for _, value := range values {
		if err := validateEventValue(reflect.ValueOf(value), 0); err != nil {
			return err
		}
	}
	return thread.fn.Call(threadEvent{name: name, values: append([]any(nil), values...)}, mode)
}

// Reject scoped JS handles inside Go containers before they leave the Node
// callback. Bound the walk so cyclic payloads cannot overflow ValueOf later.
func validateEventValue(value reflect.Value, depth int) error {
	if !value.IsValid() {
		return nil
	}
	if depth > 64 {
		return fmt.Errorf("event data is cyclic or nested more than 64 levels")
	}
	if value.Type().Implements(reflect.TypeFor[napi.ValueType]()) {
		return fmt.Errorf("event data must be Go-owned, not a JavaScript handle")
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if !value.IsNil() {
			return validateEventValue(value.Elem(), depth+1)
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			if err := validateEventValue(value.Index(i), depth+1); err != nil {
				return err
			}
		}
	case reflect.Map:
		entries := value.MapRange()
		for entries.Next() {
			if err := validateEventValue(entries.Key(), depth+1); err != nil {
				return err
			}
			if err := validateEventValue(entries.Value(), depth+1); err != nil {
				return err
			}
		}
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			if value.Type().Field(i).IsExported() {
				if err := validateEventValue(value.Field(i), depth+1); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Close drains accepted events. Abort discards them. Both cancel Run's context
// and are idempotent; the first shutdown call chooses the drain/abort mode.
// Neither waits for Done, so they may also be called from the Node thread.
func (thread *EventEmitterThread) Close() error { return thread.shutdown(napi.Release) }
func (thread *EventEmitterThread) Abort() error { return thread.shutdown(napi.Abort) }

func (thread *EventEmitterThread) shutdown(mode napi.ThreadsafeFunctionReleaseMode) error {
	thread.mu.Lock()
	if thread.closing {
		thread.mu.Unlock()
		return nil
	}
	thread.closing = true
	thread.cancel()
	thread.mu.Unlock()
	err := thread.fn.Release(mode)
	if errors.Is(err, napi.ErrThreadsafeFunctionClosing) {
		return nil
	}
	return err
}

// Done closes after native finalization and reference cleanup. Do not wait for
// it on the Node thread. Context is cancelled by shutdown or environment exit.
func (thread *EventEmitterThread) Done() <-chan struct{}    { return thread.done }
func (thread *EventEmitterThread) Context() context.Context { return thread.ctx }

// Ref/Unref control event-loop liveness and must run on the owning Node thread.
func (thread *EventEmitterThread) Ref(env napi.EnvType) error   { return thread.fn.Ref(env) }
func (thread *EventEmitterThread) Unref(env napi.EnvType) error { return thread.fn.Unref(env) }

// Run starts one processor goroutine and closes the queue when it returns.
// Errors and panics become JavaScript Error objects on the "error" event.
// Register an error listener before starting. A processor must respect context
// cancellation; Go goroutines cannot be forcibly terminated.
func (thread *EventEmitterThread) Run(process func(context.Context, *EventEmitterThread) error) error {
	if process == nil {
		return fmt.Errorf("processor cannot be nil")
	}
	thread.mu.Lock()
	if thread.closing {
		thread.mu.Unlock()
		return napi.ErrThreadsafeFunctionClosing
	}
	if thread.running {
		thread.mu.Unlock()
		return fmt.Errorf("processor has already started")
	}
	thread.running = true
	thread.mu.Unlock()
	go func() {
		defer thread.Close()
		defer func() {
			if recovered := recover(); recovered != nil {
				_ = thread.fn.Call(threadEvent{name: "error", err: fmt.Errorf("background processor panic: %v", recovered)}, napi.Blocking)
			}
		}()
		if err := process(thread.ctx, thread); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, napi.ErrThreadsafeFunctionClosing) {
			_ = thread.fn.Call(threadEvent{name: "error", err: err}, napi.Blocking)
		}
	}()
	return nil
}
