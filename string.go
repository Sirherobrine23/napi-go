package napi

import (
	"unicode/utf16"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type String struct{ value }

// Convert [ValueType] to [*String]
func ToString(o ValueType) *String { return &String{o} }

// Create [*String] from go string
func CreateString(env EnvType, str string) (*String, error) {
	napiString, err := mustValueErr(napi.CreateStringUtf8(env.NapiValue(), str))
	if err != nil {
		return nil, err
	}
	return ToString(N_APIValue(env, napiString)), nil
}

// Create string to utf16
func CreateStringUtf16(env EnvType, str []rune) (*String, error) {
	napiString, err := mustValueErr(napi.CreateStringUtf16(env.NapiValue(), utf16.Encode(str)))
	if err != nil {
		return nil, err
	}
	return ToString(N_APIValue(env, napiString)), nil
}

// Get String value.
func (str *String) Utf8Value() (string, error) {
	return mustValueErr(napi.GetValueStringUtf8(str.NapiEnv(), str.NapiValue()))
}

// Converts a String value to a UTF-16 encoded in rune.
func (str *String) Utf16Value() ([]rune, error) {
	valueOf, err := mustValueErr(napi.GetValueStringUtf16(str.NapiEnv(), str.NapiValue()))
	if err != nil {
		return nil, err
	}
	return utf16.Decode(valueOf), nil
}