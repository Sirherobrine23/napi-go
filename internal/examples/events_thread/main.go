package main

import (
	"context"
	"fmt"
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
	event "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/EventEmitter"
)

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	start, err := napi.CreateFunction(env, "start", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		count := int64(10)
		if len(ci.Args) > 0 {
			var err error
			count, err = napi.ToNumber(ci.Args[0]).Int()
			if err != nil {
				return nil, err
			}
		}
		if count < 0 || count > 10000 {
			return nil, fmt.Errorf("count must be between 0 and 10000")
		}
		thread, err := event.CreateEventThread(ci.Env, 64)
		if err != nil {
			return nil, err
		}
		emitter, err := thread.Emitter(ci.Env)
		if err != nil {
			thread.Abort()
			return nil, err
		}
		stop, err := napi.CreateFunction(ci.Env, "cancel", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return nil, thread.Abort() })
		if err != nil {
			thread.Abort()
			return nil, err
		}
		if err := napi.ToObject(emitter).Set("cancel", stop); err != nil {
			thread.Abort()
			return nil, err
		}
		if err := thread.Run(func(ctx context.Context, events *event.EventEmitterThread) error {
			var sum int64
			for index := 0; index < int(count); index++ {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				value := index * index
				sum += int64(value)
				if err := events.EmitBlocking("progress", map[string]int{"index": index, "value": value}); err != nil {
					return err
				}
			}
			return events.EmitBlocking("done", map[string]any{"count": int(count), "sum": float64(sum)})
		}); err != nil {
			thread.Abort()
			return nil, err
		}

		return emitter, nil
	})
	if err != nil {
		panic(err)
	}

	
	if err := export.Set("start", start); err != nil {
		panic(err)
	}
}

func main() {}
