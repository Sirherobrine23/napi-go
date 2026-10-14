// Package buffer wraps node:buffer's Buffer constructor. CopyBytes returns Go-owned data.
package buffer

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:buffer")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) constructor() (*napi.Object, error) {
	value, err := m.Get("Buffer")
	if err != nil {
		return nil, err
	}
	return napi.ObjectOf(m.Env(), value)
}
func (m *Module) call(method string, args ...any) (napi.ValueType, error) {
	ctor, err := m.constructor()
	if err != nil {
		return nil, err
	}
	return ctor.CallMethod(method, args...)
}
func (m *Module) From(input any, options ...any) (*napi.Buffer, error) {
	value, err := m.call("from", append([]any{input}, options...)...)
	if err != nil {
		return nil, err
	}
	return napi.ToBuffer(value), nil
}
func (m *Module) Alloc(size int, fill ...any) (*napi.Buffer, error) {
	value, err := m.call("alloc", append([]any{size}, fill...)...)
	if err != nil {
		return nil, err
	}
	return napi.ToBuffer(value), nil
}
func (m *Module) IsBuffer(value any) (bool, error) {
	result, err := m.call("isBuffer", value)
	if err != nil {
		return false, err
	}
	return napi.Decode[bool](result)
}
func (m *Module) ByteLength(input any, encoding ...string) (int, error) {
	value, err := m.call("byteLength", append([]any{input}, nodejs.Strings(encoding)...)...)
	if err != nil {
		return 0, err
	}
	return napi.Decode[int](value)
}
func (m *Module) Concat(buffers ...napi.ValueType) (*napi.Buffer, error) {
	value, err := m.call("concat", buffers)
	if err != nil {
		return nil, err
	}
	return napi.ToBuffer(value), nil
}
func (m *Module) Compare(a, b napi.ValueType) (int, error) {
	value, err := m.call("compare", a, b)
	if err != nil {
		return 0, err
	}
	return napi.Decode[int](value)
}
func ToString(value napi.ValueType, encoding ...string) (string, error) {
	result, err := napi.ToObject(value).CallMethod("toString", nodejs.Strings(encoding)...)
	if err != nil {
		return "", err
	}
	return napi.Decode[string](result)
}
func CopyBytes(value napi.ValueType) ([]byte, error) { return nodejs.CopyBytes(value) }
