package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type (
	Function     struct{ *Value }
	Callback func(env *Env, this *Value, args []*Value) (*Value, error)
)

func CreateFunction(env *Env, name string, callback Callback) (*Function, error) {
	fnCall, err := napi.MustValueErr(napi.CreateFunction(env.NapiValue(), name, func(env napi.Env, info napi.CallbackInfo) napi.Value {
		Env := FromEnvNapi(env)
		this, args, err := ReturnValuesFromCallback(env, info)
		if err != nil {
			return Env.Undefined().NapiValue()
		}

		res, err := callback(FromEnvNapi(env), this, args)
		if err != nil {
			errMsg, err2 := CreateError(Env, err.Error())
			if err2 == nil {
				errMsg.ThrowAsJavaScriptException()
				return nil
			}
			return Env.Undefined().NapiValue()
		}
		return res.NapiValue()
	}))
	if err != nil {
		return nil, err
	}

	return &Function{
		Value: &Value{
			env:     env,
			typeof:  napi.ValueTypeFunction,
			valueOf: fnCall,
		},
	}, nil
}

func (fn *Function) internalCall(Argc int, Recv *Value, Argv []napi.Value) (*Value, error) {
	res, err := napi.MustValueErr(napi.CallFunction(fn.NapiEnv(), Recv.NapiValue(), fn.NapiValue(), Argc, Argv))
	if err != nil {
		return nil, err
	}
	return &Value{
		env:     fn.env,
		valueOf: res,
		typeof:  napi.MustValue(napi.Typeof(fn.NapiEnv(), res)),
	}, nil
}

func (fn *Function) CallArgc(Argc int, Recv *Value, Argv ...*Value) (*Value, error) {
	var napiArgv []napi.Value
	for index := range Argc {
		napiArgv = append(napiArgv, Argv[index].NapiValue())
	}
	return fn.internalCall(Argc, Recv, napiArgv)
}

func (fn *Function) CallRecvArgs(recv *Value, args ...*Value) (*Value, error) {
	argc := len(args)
	stackArgsCount := 6
	stackArgs := make([]napi.Value, stackArgsCount)
	var heapArgs []napi.Value
	var argv []napi.Value
	if argc <= stackArgsCount {
		argv = stackArgs[:argc]
	} else {
		heapArgs = make([]napi.Value, argc)
		argv = heapArgs
	}
	for index := 0; index < argc; index++ {
		argv[index] = args[index].NapiValue()
	}
	return fn.internalCall(argc, recv, argv)
}

func (fn *Function) Call(args ...*Value) (*Value, error) {
	return fn.CallRecvArgs(fn.env.Undefined(), args...)
}

func ReturnValuesFromCallback(env napi.Env, info napi.CallbackInfo) (this *Value, args []*Value, err error) {
	var cbInfo napi.GetCbInfoResult
	if cbInfo, err = napi.MustValueErr(napi.GetCbInfo(env, info)); err != nil {
		return nil, nil, err
	}

	valueEnv := FromEnvNapi(env)
	this = &Value{env: valueEnv, valueOf: cbInfo.This, typeof: napi.MustValue(napi.Typeof(env, cbInfo.This))}
	args = make([]*Value, len(cbInfo.Args))
	for i, cbArg := range cbInfo.Args {
		args[i] = &Value{env: valueEnv, valueOf: cbArg, typeof: napi.MustValue(napi.Typeof(env, cbArg))}
	}
	return
}
