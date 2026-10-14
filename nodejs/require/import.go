package require

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	internalnapi "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// Import get module with require or import
func Module(env napi.EnvType, name string) (napi.ValueType, error) {
	gl, err := env.Global()
	if err != nil {
		return nil, err
	}
	exist, err := gl.Get("require")
	if err == nil && exist != nil {
		return Require(env, name)
	}
	return Import(env, name, nil)
}

func Require(env napi.EnvType, id string) (napi.ValueType, error) {
	str, err := napi.CreateString(env, id)
	if err != nil {
		return nil, err
	}
	gl, err := env.Global()
	if err != nil {
		return nil, err
	}
	return napi.ToFunction(gl.MustGet("require")).CallWithGlobal(gl, str)
}

type ImportOptions struct {
	With map[string]string `js:"with,omitempty"`
}

func Import(env napi.EnvType, id string, opt *ImportOptions) (napi.ValueType, error) {
	gl, err := env.Global()
	if err != nil {
		return nil, err
	}
	fn, err := gl.Get("import")
	str, _ := napi.CreateString(env, id)

	var res napi.ValueType
	if opt != nil {
		optionsImport, _ := napi.ValueOf(env, opt)
		res, err = napi.ToFunction(fn).Call(str, optionsImport)
	} else {
		res, err = napi.ToFunction(fn).Call(str)
	}
	if err != nil {
		return nil, err
	}

	var resType internalnapi.Value
	err = internalnapi.ResolveDeferred(env.NapiValue(), internalnapi.Deferred(res.NapiValue()), resType).ToError()
	if err == nil {
		return napi.N_APIValue(env, resType), nil
	}
	return nil, err
}
