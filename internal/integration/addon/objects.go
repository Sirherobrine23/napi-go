package main

import (
	"fmt"
	"strings"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

type objectConfig struct {
	Name    string            `napi:"name"`
	Enabled bool              `napi:"enabled"`
	Items   []int             `napi:"items"`
	Labels  map[string]string `napi:"labels"`
	Child   *objectChild      `napi:"child"`
	Hidden  string            `napi:"-"`
}

type objectChild struct {
	Score int    `napi:"score"`
	Note  string `napi:",omitempty"`
}

type objectString string
type objectBool bool
type objectError struct{}

func (objectError) Error() string { return "concrete Go error" }

func registerObjects(env napi.EnvType, export *napi.Object) {
	bind := func(name string, callback napi.Callback) {
		if err := export.SetMethod(name, callback); err != nil {
			panic(err)
		}
	}
	bind("objectCreate", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ObjectOf(ci.Env, map[string]any{"name": "Gopher", "count": 2, "nested": map[string]any{"active": true}, "items": []int{1, 2}})
		if err != nil {
			return nil, err
		}
		if err := obj.Assign(map[string]any{"ratio": float32(1.5), "unset": nil}); err != nil {
			return nil, err
		}
		if err := obj.SetValue("typedNil", (*napi.Object)(nil)); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("add", func(a, b int) int { return a + b }); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("join", func(prefix string, values ...string) string { return prefix + strings.Join(values, ",") }); err != nil {
			return nil, err
		}
		// Unnamed callback functions must receive CallbackInfo, not be decoded as structs.
		if err := obj.SetMethod("increment", func(call *napi.CallbackInfo) (napi.ValueType, error) {
			amount, err := napi.ArgAs[int](call, 0)
			if err != nil {
				return nil, err
			}
			this := napi.ToObject(call.This)
			count, err := napi.GetAs[int](this, "count")
			if err != nil {
				return nil, err
			}
			if err := this.SetValue("count", count+amount); err != nil {
				return nil, err
			}
			return napi.ValueOf(call.Env, count+amount)
		}); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("echo", func(config objectConfig) objectConfig { return config }); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("sum", func(values []int) int {
			var sum int
			for _, value := range values {
				sum += value
			}
			return sum
		}); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("nilError", func() (int, error) { return 7, nil }); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("fail", func() error { return fmt.Errorf("Go method failed") }); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("concreteError", func() objectError { return objectError{} }); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("noop", func() {}); err != nil {
			return nil, err
		}
		if err := obj.SetMethod("anyArgs", func(args ...any) []any { return args }); err != nil {
			return nil, err
		}
		return obj, nil
	})
	bind("objectDecode", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		config, err := napi.ArgAs[objectConfig](ci, 0)
		if err != nil {
			return nil, err
		}
		return napi.ObjectOf(ci.Env, config)
	})
	bind("objectAny", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		input, err := napi.ArgAs[any](ci, 0)
		if err != nil {
			return nil, err
		}
		return napi.ValueOf(ci.Env, input)
	})
	bind("objectGetString", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ArgAs[*napi.Object](ci, 0)
		if err != nil {
			return nil, err
		}
		name, err := napi.ArgAs[string](ci, 1)
		if err != nil {
			return nil, err
		}
		value, err := napi.GetAs[objectString](obj, name)
		if err != nil {
			return nil, err
		}
		return napi.ValueOf(ci.Env, value)
	})
	bind("objectGetWrapper", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ArgAs[*napi.Object](ci, 0)
		if err != nil {
			return nil, err
		}
		return napi.GetAs[*napi.String](obj, "name")
	})
	bind("objectGetBool", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ArgAs[*napi.Object](ci, 0)
		if err != nil {
			return nil, err
		}
		var flag objectBool
		if err := obj.GetInto("flag", &flag); err != nil {
			return nil, err
		}
		return napi.ValueOf(ci.Env, flag)
	})
	bind("objectCall", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ArgAs[*napi.Object](ci, 0)
		if err != nil {
			return nil, err
		}
		return obj.CallMethod("method", "from Go", map[string]any{"count": 3})
	})
	bind("objectCallZero", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ArgAs[*napi.Object](ci, 0)
		if err != nil {
			return nil, err
		}
		value, err := obj.CallMethod("method")
		if err != nil {
			return nil, err
		}
		return napi.GetAs[*napi.Number](must(napi.ObjectOf(ci.Env, map[string]any{"result": value})), "result")
	})
	bind("objectFunctionCall", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn, err := napi.ArgAs[*napi.Function](ci, 0)
		if err != nil {
			return nil, err
		}
		value, err := fn.CallValues(4, []int{5, 6})
		if err != nil {
			return nil, err
		}
		result, err := napi.Decode[int](value)
		if err != nil {
			return nil, err
		}
		return napi.ValueOf(ci.Env, result)
	})
	bind("objectFunctionThis", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		fn, err := napi.ArgAs[*napi.Function](ci, 0)
		if err != nil {
			return nil, err
		}
		obj, err := napi.ArgAs[*napi.Object](ci, 1)
		if err != nil {
			return nil, err
		}
		return fn.CallWithThis(obj, 3)
	})
	bind("objectClassNew", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		class, err := napi.ArgAs[*napi.Class](ci, 0)
		if err != nil {
			return nil, err
		}
		instance, err := class.NewValues("Gopher", 23)
		if err != nil {
			return nil, err
		}
		return instance, nil
	})
	bind("objectUnwrap", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ArgAs[*napi.Object](ci, 0)
		if err != nil {
			return nil, err
		}
		state, err := napi.UnwrapAs[*counter](obj)
		if err != nil {
			return nil, err
		}
		return napi.ValueOf(ci.Env, state.value)
	})
	bind("objectUnwrapMismatch", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		_, err := napi.UnwrapAs[*objectChild](napi.ToObject(ci.Args[0]))
		return nil, err
	})
	bind("objectInvalid", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		kind, err := napi.ArgAs[string](ci, 0)
		if err != nil {
			return nil, err
		}
		obj := must(napi.CreateObject(ci.Env))
		switch kind {
		case "primitive":
			return napi.ObjectOf(ci.Env, 23)
		case "nil-function":
			return nil, obj.SetMethod("invalid", (func())(nil))
		case "not-function":
			return nil, obj.SetMethod("invalid", 2)
		case "nil-target":
			return nil, napi.ValueFrom(must(napi.ValueOf(ci.Env, "x")), (*string)(nil))
		case "value-target":
			return nil, napi.ValueFrom(must(napi.ValueOf(ci.Env, "x")), "x")
		case "nil-value":
			return nil, obj.Set("invalid", (*napi.Object)(nil))
		case "negative-argument":
			var target int
			return nil, ci.ArgInto(-1, &target)
		case "cyclic":
			cyclic := map[string]any{}
			cyclic["self"] = cyclic
			return nil, obj.SetValue("cyclic", cyclic)
		case "foreign":
			// A synthetic env is used only to exercise rejection before any native
			// call. No value from a different JS thread is read here.
			foreign := napi.N_APIValue(napi.N_APIEnv(nil), obj.NapiValue())
			return nil, obj.SetValue("foreign", map[string]any{"value": foreign})
		default:
			return nil, fmt.Errorf("unknown validation case")
		}
	})
	bind("objectAssign", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ArgAs[*napi.Object](ci, 0)
		if err != nil {
			return nil, err
		}
		if err := obj.Assign(map[string]any{"b": 2, "a": 1}); err != nil {
			return nil, err
		}
		return obj, nil
	})
	bind("objectAssignInvalid", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		obj, err := napi.ArgAs[*napi.Object](ci, 0)
		if err != nil {
			return nil, err
		}
		cyclic := map[string]any{}
		cyclic["self"] = cyclic
		return nil, obj.Assign(map[string]any{"a": 1, "b": cyclic, "c": 3})
	})
}
