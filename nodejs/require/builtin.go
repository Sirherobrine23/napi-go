package require

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

// Builtin loads a Node builtin synchronously, including from ESM and Workers.
// Modern Node provides process.getBuiltinModule; older hosts can expose a
// synchronous global require. No Promise is treated as a native deferred value.
func Builtin(env napi.EnvType, id string) (napi.ValueType, error) {
	global, err := env.Global()
	if err != nil {
		return nil, err
	}
	process, err := global.Get("process")
	if err != nil {
		return nil, err
	}
	getter, err := napi.ToObject(process).Get("getBuiltinModule")
	if err != nil {
		return nil, err
	}
	typeOf, err := getter.Type()
	if err != nil {
		return nil, err
	}
	if typeOf == napi.TypeFunction {
		name, err := napi.CreateString(env, id)
		if err != nil {
			return nil, err
		}
		module, err := napi.ToFunction(getter).CallWithGlobal(process, name)
		if err != nil {
			return nil, err
		}
		kind, err := module.Type()
		if err != nil {
			return nil, err
		}
		if kind == napi.TypeUndefined {
			return nil, fmt.Errorf("unknown Node builtin %q", id)
		}
		return module, nil
	}
	loader, err := global.Get("require")
	if err != nil {
		return nil, err
	}
	typeOf, err = loader.Type()
	if err != nil {
		return nil, err
	}
	if typeOf != napi.TypeFunction {
		return nil, fmt.Errorf("loading %q requires process.getBuiltinModule or a global require; pass an existing EventEmitter instead", id)
	}
	return Require(env, id)
}
