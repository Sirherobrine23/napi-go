// Package util exposes Node formatting, inspection and deep equality helpers.
package util

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:util")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) Format(args ...any) (string, error) {
	return nodejs.CallAs[string](m.Module, "format", args...)
}
func (m *Module) Inspect(value any, options ...any) (string, error) {
	return nodejs.CallAs[string](m.Module, "inspect", append([]any{value}, options...)...)
}
func (m *Module) IsDeepStrictEqual(a, b any) (bool, error) {
	return nodejs.CallAs[bool](m.Module, "isDeepStrictEqual", a, b)
}
func (m *Module) Promisify(function napi.ValueType) (*napi.Function, error) {
	return nodejs.CallAs[*napi.Function](m.Module, "promisify", function)
}
func (m *Module) Callbackify(function napi.ValueType) (*napi.Function, error) {
	return nodejs.CallAs[*napi.Function](m.Module, "callbackify", function)
}
