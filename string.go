package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type String struct{ value }

func FromValue(value ValueType) *String { return &String{value: value} }

func CreateString(env EnvType, value string) (*String, error) {
	napiValue, err := napi.MustValueErr(napi.CreateStringUtf8(env.NapiValue(), value))
	if err != nil {
		return nil, err
	}
	return &String{value: FromValueNapi(env, napiValue)}, nil
}

func MustCreateString(env EnvType, value string) *String {
	valueType, err := CreateString(env, value)
	if err != nil {
		panic(err)
	}
	return valueType
}

func (str *String) ValueOf() (string, error) {
	return napi.MustValueErr(napi.GetValueStringUtf8(str.NapiEnv(), str.NapiValue()))
}

func (str *String) String() string {
	stri, err := str.ValueOf()
	if err != nil {
		return ""
	}
	return stri
}
