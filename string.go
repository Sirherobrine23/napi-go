package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type String struct{ *Value }

func FromValue(value *Value) *String { return &String{Value: value} }

func MustCreateString(env *Env, value string) *String {
	napiValue, status := napi.CreateStringUtf8(env.env, value)
	if status != napi.StatusOK {
		panic(napi.StatusError(status))
	}
	return &String{
		Value: &Value{
			env:     env,
			typeof:  napi.ValueTypeString,
			valueOf: napiValue,
		},
	}
}

func CreateString(env *Env, value string) (*String, error) {
	napiValue, status := napi.CreateStringUtf8(env.env, value)
	if status != napi.StatusOK {
		return nil, napi.StatusError(status)
	}
	return &String{
		Value: &Value{
			env:     env,
			typeof:  napi.ValueTypeString,
			valueOf: napiValue,
		},
	}, nil
}

func (str *String) String() string {
	if str != nil {
		return napi.MustValue(napi.GetValueStringUtf8(str.NapiEnv(), str.valueOf))
	}
	return ""
}
