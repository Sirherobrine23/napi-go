package napi

import (
	"fmt"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Function struct{ value }

// Function to call on Javascript caller
type Callback func(env EnvType, this ValueType, args []ValueType) (ValueType, error)

// Convert [ValueType] to [*Function]
func ToFunction(o ValueType) *Function { return &Function{o} }

func CreateFunction(env EnvType, name string, callback Callback) (*Function, error) {
	return CreateFunctionNapi(env, name, func(napiEnv napi.Env, info napi.CallbackInfo) napi.Value {
		env := N_APIEnv(napiEnv)
		cbInfo, err := mustValueErr(napi.GetCbInfo(napiEnv, info))
		if err != nil {
			ThrowError(env, "", err.Error())
			return nil
		}

		this := N_APIValue(env, cbInfo.This)
		args := make([]ValueType, len(cbInfo.Args))
		for i, cbArg := range cbInfo.Args {
			args[i] = N_APIValue(env, cbArg)
		}

		defer func() {
			if err := recover(); err != nil {
				switch v := err.(type) {
				case error:
					ThrowError(env, "", v.Error())
				default:
					ThrowError(env, "", fmt.Sprintf("panic recover: %s", err))
				}
			}
		}()

		res, err := callback(env, this, args)
		switch {
		case err != nil:
			ThrowError(env, "", err.Error())
			return nil
		case res == nil:
			und, _ := env.Undefined()
			return und.NapiValue()
		default:
			typeOf, _ := res.Type()
			if typeOf == TypeError {
				ToError(res).ThrowAsJavaScriptException()
				return nil
			}
			return res.NapiValue()
		}
	})
}

func CreateFunctionNapi(env EnvType, name string, callback napi.Callback) (*Function, error) {
	fnCall, err := mustValueErr(napi.CreateFunction(env.NapiValue(), name, callback))
	if err != nil {
		return nil, err
	}
	return ToFunction(N_APIValue(env, fnCall)), nil
}

func (fn *Function) internalCall(this napi.Value, argc int, argv []napi.Value) (ValueType, error) {
	// napi_call_function(env, global, add_two, argc, argv, &return_val);
	res, err := mustValueErr(napi.CallFunction(fn.NapiEnv(), this, fn.NapiValue(), argc, argv))
	if err != nil {
		return nil, err
	}
	return N_APIValue(fn.Env(), res), nil
}

// Call function with custom global/this value
func (fn *Function) CallWithGlobal(this ValueType, args ...ValueType) (ValueType, error) {
	argc := len(args)
	argv := make([]napi.Value, argc)
	for index := range argc {
		argv[index] = args[index].NapiValue()
	}
	return fn.internalCall(this.NapiValue(), argc, argv)
}

// Call function with args
func (fn *Function) Call(args ...ValueType) (ValueType, error) {
	global, err := fn.Env().Global()
	if err != nil {
		return nil, err
	}
	return fn.CallWithGlobal(global, args...)
}
