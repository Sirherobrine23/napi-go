package main

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
)

var finalized, delivered, collected, wrapped atomic.Int32

type counter struct{ value int }
type payload struct{ value int }

func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func state(ci *napi.CallbackInfo) *counter {
	return must(napi.ToObject(ci.This).Unwrap()).(*counter)
}

func counterClass(env napi.EnvType) *napi.Class {
	return must(napi.CreateClass(env, "Counter", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		c := &counter{}
		if len(ci.Args) > 0 {
			c.value = int(must(napi.ToNumber(ci.Args[0]).Int()))
		}
		return nil, napi.ToObject(ci.This).Wrap(c, func(data any) { wrapped.Add(1) })
	},
		napi.PropertyDescriptor{Utf8name: "increment", Attributes: napi.DefaultMethod, Method: func(ci *napi.CallbackInfo) (napi.ValueType, error) {
			state(ci).value++
			return napi.CreateNumber(ci.Env, state(ci).value)
		}},
		napi.PropertyDescriptor{Utf8name: "value", Attributes: napi.Enumerable | napi.Configurable,
			Getter: func(ci *napi.CallbackInfo) (napi.ValueType, error) { return napi.CreateNumber(ci.Env, state(ci).value) },
			Setter: func(ci *napi.CallbackInfo) (napi.ValueType, error) {
				state(ci).value = int(must(napi.ToNumber(ci.Args[0]).Int()))
				return nil, nil
			}},
		napi.PropertyDescriptor{Utf8name: "label", Value: must(napi.CreateString(env, "counter")), Attributes: napi.Enumerable},
		napi.PropertyDescriptor{Utf8name: "version", Attributes: napi.Static | napi.DefaultMethod, Method: func(ci *napi.CallbackInfo) (napi.ValueType, error) { return napi.CreateString(ci.Env, "1") }},
		napi.PropertyDescriptor{Utf8name: "fail", Method: func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, fmt.Errorf("class error") }},
		napi.PropertyDescriptor{Utf8name: "panic", Method: func(ci *napi.CallbackInfo) (napi.ValueType, error) { panic("class panic") }},
	))
}

func control(env napi.EnvType, fn *napi.ThreadsafeFunction) *napi.Object {
	obj := must(napi.CreateObject(env))
	obj.Set("closed", must(napi.CreateFunction(env, "closed", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		_, err := fn.GetContext()
		return napi.CreateBoolean(ci.Env, errors.Is(err, napi.ErrThreadsafeFunctionClosing))
	})))
	obj.Set("rejects", must(napi.CreateFunction(env, "rejects", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		return napi.CreateBoolean(ci.Env, errors.Is(fn.Call(1, napi.NonBlocking), napi.ErrThreadsafeFunctionClosing) && errors.Is(fn.Acquire(), napi.ErrThreadsafeFunctionClosing) && errors.Is(fn.Release(napi.Release), napi.ErrThreadsafeFunctionClosing))
	})))
	return obj
}

func createFn(env napi.EnvType, js any, callback napi.ThreadsafeFunctionCallJSCallback, queue int, final bool) *napi.ThreadsafeFunction {
	var finalize napi.ThreadsafeFunctionFinalizeCallback
	if final {
		finalize = func(env napi.EnvType, context any) {
			if context != "context" {
				panic("context corrupted")
			}
			finalized.Add(1)
		}
	}
	return must(napi.CreateThreadsafeFunction(env, js, finalize, callback, "integration", queue, 1, "context"))
}

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	registerBuiltins(env, export)
	registerObjects(env, export)
	registerEvents(env, export)
	class := counterClass(env)
	export.Set("Counter", class)
	export.Set("zero", must(class.New()))
	export.Set("one", must(class.New(must(napi.CreateNumber(env, 1)))))
	export.Set("Empty", must(napi.CreateClass(env, "Empty", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, nil })))
	bind := func(name string, callback napi.Callback) {
		if err := export.Set(name, must(napi.CreateFunction(env, name, callback))); err != nil {
			panic(err)
		}
	}
	registerStatusErrors(env, export)
	registerEventThreads(env, export)
	bind("makeClass", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return counterClass(ci.Env), nil })
	bind("detach", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		data := must(napi.ToObject(ci.Args[0]).RemoveWrap()).(*counter)
		return napi.CreateNumber(ci.Env, data.value)
	})
	bind("unwrap", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		_, err := napi.ToObject(ci.Args[0]).Unwrap()
		return nil, err
	})
	bind("rewrap", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		return nil, napi.ToObject(ci.Args[0]).Wrap(&counter{42}, nil)
	})
	bind("properties", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj := must(napi.CreateObject(ci.Env))
		if err := obj.DefineProperties(); err != nil {
			return nil, err
		}
		key := ci.Args[0]
		return obj, obj.DefineProperties(napi.PropertyDescriptor{Name: key, Value: must(napi.CreateNumber(ci.Env, 23)), Attributes: napi.DefaultJSProperty},
			napi.PropertyDescriptor{Utf8name: "method", Method: func(ci *napi.CallbackInfo) (napi.ValueType, error) { return napi.CreateNumber(ci.Env, 7) }},
			napi.PropertyDescriptor{Utf8name: "getter", Getter: func(ci *napi.CallbackInfo) (napi.ValueType, error) { return napi.CreateNumber(ci.Env, 8) }})
	})
	bind("stream", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		count := int(must(napi.ToNumber(ci.Args[1]).Int()))
		producers := int(must(napi.ToNumber(ci.Args[2]).Int()))
		fn := createFn(ci.Env, napi.As[*napi.Function](ci.Args[0]), nil, 2, true)
		for i := 0; i < producers; i++ {
			if err := fn.Acquire(); err != nil {
				panic(err)
			}
		}
		if err := fn.Release(napi.Release); err != nil {
			panic(err)
		}
		for producer := 0; producer < producers; producer++ {
			go func(producer int) {
				defer fn.Release(napi.Release)
				for i := producer; i < count; i += producers {
					if err := fn.Call(i, napi.Blocking); err != nil {
						return
					}
				}
			}(producer)
		}
		return control(ci.Env, fn), nil
	})
	bind("queue", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, ci.Args[0], nil, 1, true)
		if err := fn.Call(1, napi.NonBlocking); err != nil {
			panic(err)
		}
		if !errors.Is(fn.Call(2, napi.NonBlocking), napi.ErrThreadsafeFunctionQueueFull) {
			panic("queue not full")
		}
		if err := fn.Release(napi.Release); err != nil {
			panic(err)
		}
		return control(ci.Env, fn), nil
	})
	bind("custom", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		var fn *napi.ThreadsafeFunction
		fn = createFn(ci.Env, nil, func(env napi.EnvType, js *napi.Function, data any) {
			if js != nil || data.(string) != "data" || must(fn.GetContext()) != "context" {
				panic("custom callback data corrupted")
			}
			delivered.Add(1)
		}, 0, true)
		if err := fn.Call("data", napi.NonBlocking); err != nil {
			panic(err)
		}
		fn.Release(napi.Release)
		return control(ci.Env, fn), nil
	})
	bind("abort", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, nil, func(env napi.EnvType, js *napi.Function, data any) { delivered.Add(1) }, 0, true)
		for i := 0; i < 1000; i++ {
			p := &payload{i}
			runtime.SetFinalizer(p, func(*payload) { collected.Add(1) })
			if err := fn.Call(p, napi.NonBlocking); err != nil {
				panic(err)
			}
		}
		fn.Release(napi.Abort)
		return control(ci.Env, fn), nil
	})
	bind("abortBlocking", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, nil, func(env napi.EnvType, js *napi.Function, data any) {}, 1, true)
		if err := fn.Call(1, napi.NonBlocking); err != nil {
			panic(err)
		}
		var started sync.WaitGroup
		started.Add(4)
		done := make(chan error, 4)
		for i := 0; i < 4; i++ {
			if err := fn.Acquire(); err != nil {
				panic(err)
			}
			go func() {
				defer fn.Release(napi.Release)
				started.Done()
				done <- fn.Call(2, napi.Blocking)
			}()
		}
		started.Wait()
		time.Sleep(20 * time.Millisecond)
		if err := fn.Release(napi.Abort); err != nil {
			panic(err)
		}
		for i := 0; i < 4; i++ {
			select {
			case err := <-done:
				if !errors.Is(err, napi.ErrThreadsafeFunctionClosing) {
					panic("blocked call did not close")
				}
			case <-time.After(time.Second):
				panic("Abort deadlocked with blocking producers")
			}
		}
		return control(ci.Env, fn), nil
	})
	bind("noFinalizer", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, ci.Args[0], nil, 0, false)
		fn.Call(nil, napi.NonBlocking)
		fn.Release(napi.Release)
		return control(ci.Env, fn), nil
	})
	bind("goCallback", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, func(ci *napi.CallbackInfo) (napi.ValueType, error) { delivered.Add(1); return nil, nil }, nil, 0, true)
		fn.Call(nil, napi.NonBlocking)
		fn.Release(napi.Release)
		return control(ci.Env, fn), nil
	})
	bind("converted", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, ci.Args[0], nil, 0, true)
		for _, data := range []any{uint(23), float32(1.5), map[string]any{"value": 5, "nil": nil}} {
			if err := fn.Call(data, napi.NonBlocking); err != nil {
				panic(err)
			}
		}
		fn.Release(napi.Release)
		return control(ci.Env, fn), nil
	})
	bind("validation", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		callback := func(env napi.EnvType, js *napi.Function, data any) {}
		for _, args := range [][3]int{{-1, 1, 0}, {0, 0, 0}} {
			if _, err := napi.CreateThreadsafeFunction(ci.Env, nil, nil, callback, "bad", args[0], args[1], nil); err == nil {
				panic("accepted invalid sizes")
			}
		}
		if _, err := napi.CreateThreadsafeFunction(ci.Env, nil, nil, nil, "bad", 0, 1, nil); err == nil {
			panic("accepted missing callbacks")
		}
		if _, err := napi.CreateThreadsafeFunction(ci.Env, ci.Args[0], nil, nil, "bad", 0, 1, nil); err == nil {
			panic("accepted non-function")
		}
		if _, err := napi.CreateClass(ci.Env, "Bad", nil); err == nil {
			panic("accepted nil constructor")
		}
		if _, err := napi.CreateClass(ci.Env, "Bad", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, nil }, napi.PropertyDescriptor{Utf8name: "x", Method: func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, nil }, Value: ci.Args[0]}); err == nil {
			panic("accepted ambiguous property")
		}
		fn := createFn(ci.Env, nil, callback, 0, false)
		if fn.Call(ci.Args[0], napi.NonBlocking) == nil || fn.Call(0, napi.ThreadsafeFunctionCallMode(99)) == nil || fn.Release(napi.ThreadsafeFunctionReleaseMode(99)) == nil {
			panic("accepted invalid calls")
		}
		if err := fn.Unref(ci.Env); err != nil {
			panic(err)
		}
		if err := fn.Unref(ci.Env); err != nil {
			panic(err)
		}
		if err := fn.Ref(ci.Env); err != nil {
			panic(err)
		}
		if err := fn.Ref(ci.Env); err != nil {
			panic(err)
		}
		fn.Release(napi.Release)
		return nil, nil
	})
	bind("unref", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, nil, func(env napi.EnvType, js *napi.Function, data any) {}, 0, false)
		for i := 0; i < 1000; i++ {
			fn.Call(i, napi.NonBlocking)
		}
		return nil, fn.Unref(ci.Env)
	})
	bind("workerActive", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, nil, func(env napi.EnvType, js *napi.Function, data any) {}, 1, false)
		if err := fn.Unref(ci.Env); err != nil {
			return nil, err
		}
		go func() {
			defer fn.Release(napi.Release)
			for i := 0; i < 100000; i++ {
				if err := fn.Call(i, napi.Blocking); err != nil {
					return
				}
			}
		}()
		return nil, nil
	})
	bind("stats", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		runtime.GC()
		return napi.ValueOf(ci.Env, map[string]int32{"finalized": finalized.Load(), "delivered": delivered.Load(), "collected": collected.Load(), "wrapped": wrapped.Load()})
	})
	// Concurrent acquires/releases are also exercised without queue traffic.
	bind("acquisitions", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn := createFn(ci.Env, nil, func(env napi.EnvType, js *napi.Function, data any) {}, 0, true)
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := 0; j < 100; j++ {
					if err := fn.Acquire(); err != nil {
						panic(err)
					}
					if err := fn.Release(napi.Release); err != nil {
						panic(err)
					}
				}
			}()
		}
		wg.Wait()
		fn.Release(napi.Release)
		return control(ci.Env, fn), nil
	})
}

func main() {}
