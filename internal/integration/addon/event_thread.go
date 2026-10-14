package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	eventpkg "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/EventEmitter"
)

func exposeThread(env napi.EnvType, thread *eventpkg.EventEmitterThread) *napi.Object {
	control := must(napi.CreateObject(env))
	bind := func(name string, callback napi.Callback) {
		if err := control.Set(name, must(napi.CreateFunction(env, name, callback))); err != nil {
			panic(err)
		}
	}
	bind("emitter", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return thread.Emitter(ci.Env) })
	bind("closed", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		select {
		case <-thread.Done():
			return napi.CreateBoolean(ci.Env, true)
		default:
			return napi.CreateBoolean(ci.Env, false)
		}
	})
	bind("close", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, thread.Close() })
	bind("abort", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, thread.Abort() })
	bind("emit", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, thread.Emit("manual", "data") })
	bind("unref", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, thread.Unref(ci.Env) })
	bind("ref", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, thread.Ref(ci.Env) })
	return control
}

func registerEventThreads(env napi.EnvType, export *napi.Object) {
	bind := func(name string, callback napi.Callback) {
		if err := export.Set(name, must(napi.CreateFunction(env, name, callback))); err != nil {
			panic(err)
		}
	}
	bind("nativeEmitter", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return eventpkg.CreateEvent(ci.Env) })
	bind("nativeListen", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		emitter := eventpkg.ToEventEmitter(ci.Args[0])
		method := must(napi.ToString(ci.Args[1]).Utf8Value())
		name := must(napi.ToString(ci.Args[2]).Utf8Value())
		callback := napi.ToFunction(ci.Args[3])
		switch method {
		case "on":
			return nil, emitter.On(name, callback)
		case "once":
			return nil, emitter.Once(name, callback)
		case "prepend":
			return nil, emitter.PrependListener(name, callback)
		case "prependOnce":
			return nil, emitter.PrependOnceListener(name, callback)
		case "remove":
			return nil, emitter.RemoveListener(name, callback)
		}
		return nil, fmt.Errorf("unknown method")
	})
	bind("nativeEmit", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		ok, err := eventpkg.ToEventEmitter(ci.Args[0]).Emit(must(napi.ToString(ci.Args[1]).Utf8Value()), ci.Args[2:]...)
		if err != nil {
			return nil, err
		}
		return napi.CreateBoolean(ci.Env, ok)
	})
	bind("nativeInfo", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		emitter := eventpkg.ToEventEmitter(ci.Args[0])
		if err := emitter.SetMaxListeners(23); err != nil {
			return nil, err
		}
		names := must(emitter.EventNames())
		strings := make([]string, len(names))
		for i, name := range names {
			strings[i] = must(name.Utf8Value())
		}
		count := must(emitter.ListenerCount("native"))
		listeners := must(emitter.Listeners("native"))
		return napi.ValueOf(ci.Env, map[string]any{"names": strings, "count": float64(count), "listeners": len(listeners), "max": float64(must(emitter.GetMaxListeners()))})
	})
	bind("nativeRemoveAll", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		return nil, eventpkg.ToEventEmitter(ci.Args[0]).RemoveAllListeners()
	})
	bind("threadManual", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		var thread *eventpkg.EventEmitterThread
		if len(ci.Args) > 0 {
			thread = must(eventpkg.ToEventEmitter(ci.Args[0]).Thread(1))
		} else {
			thread = must(eventpkg.CreateEventThread(ci.Env, 1))
		}
		return exposeThread(ci.Env, thread), nil
	})
	bind("threadRun", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		mode := must(napi.ToString(ci.Args[0]).Utf8Value())
		thread := must(eventpkg.CreateEventThread(ci.Env, 2))
		emitter := must(thread.Emitter(ci.Env))
		obj := exposeThread(ci.Env, thread)
		obj.Set("value", emitter)
		err := thread.Run(func(ctx context.Context, emitter *eventpkg.EventEmitterThread) error {
			switch mode {
			case "error":
				return errors.New("processor failed")
			case "panic":
				panic("processor panic")
			case "wait":
				<-ctx.Done()
				return ctx.Err()
			case "active":
				for i := 0; i < 100000; i++ {
					if err := emitter.EmitBlocking("progress", i); err != nil {
						return err
					}
				}
			case "concurrent":
				var wg sync.WaitGroup
				for producer := 0; producer < 4; producer++ {
					wg.Add(1)
					go func(producer int) {
						defer wg.Done()
						for i := producer; i < 96; i += 4 {
							if err := emitter.EmitBlocking("progress", i); err != nil {
								return
							}
						}
					}(producer)
				}
				wg.Wait()
			default:
				for i := 0; i < 32; i++ {
					if err := emitter.EmitBlocking("progress", i, map[string]int{"square": i * i}); err != nil {
						return err
					}
				}
			}
			return emitter.EmitBlocking("done")
		})
		if err != nil {
			thread.Abort()
			return nil, err
		}
		return obj, nil
	})
	bind("threadQueue", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		thread := must(eventpkg.CreateEventThread(ci.Env, 1))
		obj := exposeThread(ci.Env, thread)
		obj.Set("value", must(thread.Emitter(ci.Env)))
		if err := thread.Emit("queued", 1); err != nil {
			panic(err)
		}
		if !errors.Is(thread.Emit("queued", 2), napi.ErrThreadsafeFunctionQueueFull) {
			panic("queue-full error lost")
		}
		if thread.Emit("bad", ci.Args[0]) == nil {
			panic("JS handle accepted as Go data")
		}
		if thread.Emit("bad", map[string]any{"nested": ci.Args[0]}) == nil {
			panic("nested JS handle accepted as Go data")
		}
		cycle := map[string]any{}
		cycle["self"] = cycle
		if thread.Emit("bad", cycle) == nil {
			panic("cyclic payload accepted")
		}
		thread.Close()
		return obj, nil
	})
	bind("threadValidation", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		if _, err := eventpkg.CreateEventThread(ci.Env, -1); err == nil {
			panic("negative queue accepted")
		}
		thread := must(eventpkg.CreateEventThread(ci.Env, 1))
		if _, err := thread.Emitter(nil); err == nil {
			panic("nil emitter environment accepted")
		}
		if thread.Run(nil) == nil {
			panic("nil processor accepted")
		}
		if err := thread.Run(func(ctx context.Context, thread *eventpkg.EventEmitterThread) error { <-ctx.Done(); return ctx.Err() }); err != nil {
			panic(err)
		}
		if thread.Run(func(context.Context, *eventpkg.EventEmitterThread) error { return nil }) == nil {
			panic("second processor accepted")
		}
		thread.Abort()
		if !errors.Is(thread.Run(func(context.Context, *eventpkg.EventEmitterThread) error { return nil }), napi.ErrThreadsafeFunctionClosing) {
			panic("closed processor accepted")
		}
		return nil, nil
	})
	bind("threadManualError", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		thread := must(eventpkg.CreateEventThread(ci.Env, 1))
		obj := exposeThread(ci.Env, thread)
		obj.Set("value", must(thread.Emitter(ci.Env)))
		if err := thread.Emit("error", errors.New("manual error")); err != nil {
			thread.Abort()
			return nil, err
		}
		thread.Close()
		return obj, nil
	})
	bind("threadAfterScope", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		thread := must(eventpkg.CreateEventThread(ci.Env, 1))
		obj := exposeThread(ci.Env, thread)
		// Block Go until a later callback starts processing; no initial emitter
		// handle is retained. Emitter(env) must resolve its persistent reference.
		start := must(napi.CreateFunction(ci.Env, "start", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
			return nil, thread.Run(func(context.Context, *eventpkg.EventEmitterThread) error {
				time.Sleep(time.Millisecond)
				return thread.EmitBlocking("done", 42)
			})
		}))
		obj.Set("start", start)
		return obj, nil
	})
}
