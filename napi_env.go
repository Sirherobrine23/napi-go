package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type EnvType interface {
	NapiValue() napi.Env // Primitive value to NAPI call
	Global() (*Object, error)
	Undefined() (ValueType, error)
	Null() (ValueType, error)
}

// Return N-API env reference
func N_APIEnv(env napi.Env) EnvType { return &Env{env} }

// N-API Env
type Env struct {
	NapiEnv napi.Env
}

// Return [napi.Env] to point from internal napi cgo
func (e *Env) NapiValue() napi.Env {
	return e.NapiEnv
}

// Return representantion to 'This' [*Object]
func (e *Env) Global() (*Object, error) {
	napiValue, err := mustValueErr(napi.GetGlobal(e.NapiEnv))
	if err != nil {
		return nil, err
	}
	return ToObject(N_APIValue(e, napiValue)), nil
}

// Return Undefined value
func (e *Env) Undefined() (ValueType, error) {
	napiValue, err := mustValueErr(napi.GetUndefined(e.NapiEnv))
	if err != nil {
		return nil, err
	}
	return N_APIValue(e, napiValue), nil
}

// Return Null value
func (e *Env) Null() (ValueType, error) {
	napiValue, err := mustValueErr(napi.GetNull(e.NapiEnv))
	if err != nil {
		return nil, err
	}
	return N_APIValue(e, napiValue), nil
}
