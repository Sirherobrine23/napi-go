package main

import (
	"fmt"
	"sync"
	"sync/atomic"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

// Only immutable Go data crosses the queue. The default TSFN callback converts
// this envelope on the Node thread; JavaScript relays it to emitter.emit.
type event struct {
	Name string `napi:"name"`
	Args []any  `napi:"args"`
}

// Used with emitter.emit.bind(emitter), without a JavaScript envelope relay.
func dispatchEvent(env napi.EnvType, callback *napi.Function, data any) {
	message := data.(event)
	args := []napi.ValueType{must(napi.CreateString(env, message.Name))}
	for _, arg := range message.Args {
		args = append(args, must(napi.ValueOf(env, arg)))
	}
	if _, err := callback.Call(args...); err != nil {
		panic(err)
	}
}

func startEvents(ci *napi.CallbackInfo, bound bool) (napi.ValueType, error) {
	if len(ci.Args) != 3 {
		return nil, fmt.Errorf("events requires a callback, item count and producer count")
	}
	count, err := napi.ToNumber(ci.Args[1]).Int()
	if err != nil {
		return nil, err
	}
	producers, err := napi.ToNumber(ci.Args[2]).Int()
	if err != nil {
		return nil, err
	}
	if count < 0 || count > 10000 || producers < 1 || producers > 32 {
		return nil, fmt.Errorf("invalid item count or producer count")
	}
	var callback napi.ThreadsafeFunctionCallJSCallback
	if bound {
		callback = dispatchEvent
	}
	fn := createFn(ci.Env, ci.Args[0], callback, 2, true)
	var coordinatorReleased atomic.Bool
	releaseCoordinator := func(mode napi.ThreadsafeFunctionReleaseMode) error {
		if !coordinatorReleased.CompareAndSwap(false, true) {
			return napi.ErrThreadsafeFunctionClosing
		}
		return fn.Release(mode)
	}
	obj := control(ci.Env, fn)
	if err := obj.Set("cancel", must(napi.CreateFunction(ci.Env, "cancel", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		return nil, releaseCoordinator(napi.Abort)
	}))); err != nil {
		fn.Release(napi.Abort)
		return nil, err
	}
	var wg sync.WaitGroup
	var total atomic.Int64
	for producer := 0; producer < int(producers); producer++ {
		if err := fn.Acquire(); err != nil {
			fn.Release(napi.Abort)
			return nil, err
		}
		wg.Add(1)
		go func(producer int) {
			defer wg.Done()
			defer fn.Release(napi.Release)
			for index := producer; index < int(count); index += int(producers) {
				value := index * index
				total.Add(int64(value))
				message := event{"progress", []any{index, map[string]int{"producer": producer, "value": value}}}
				if err := fn.Call(message, napi.Blocking); err != nil {
					return
				}
			}
		}(producer)
	}
	// The initial acquisition belongs to this coordinator, which emits done
	// only after every producer enqueued its progress events and stopped.
	go func() {
		defer releaseCoordinator(napi.Release)
		wg.Wait()
		_ = fn.Call(event{"done", []any{map[string]any{"count": int(count), "sum": float64(total.Load())}}}, napi.Blocking)
	}()
	return obj, nil
}

func registerEvents(env napi.EnvType, export *napi.Object) {
	bind := func(name string, callback napi.Callback) {
		if err := export.Set(name, must(napi.CreateFunction(env, name, callback))); err != nil {
			panic(err)
		}
	}
	bind("events", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return startEvents(ci, false) })
	bind("eventsBound", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return startEvents(ci, true) })
	bind("eventFailure", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		if len(ci.Args) != 1 {
			return nil, fmt.Errorf("eventFailure requires a callback")
		}
		fn := createFn(ci.Env, ci.Args[0], nil, 2, true)
		go func() {
			defer fn.Release(napi.Release)
			_ = fn.Call(event{"error", []any{"processing failed"}}, napi.Blocking)
		}()
		return control(ci.Env, fn), nil
	})
}
