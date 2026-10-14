// Package zlib provides synchronous Node compression. These calls block the Node
// thread; async and stream variants remain available through Module.Call/New.
package zlib

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:zlib")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) transform(method string, input any, options []any) ([]byte, error) {
	data, err := nodejs.Binary(m.Env(), input)
	if err != nil {
		return nil, err
	}
	result, err := m.Call(method, append([]any{data}, options...)...)
	if err != nil {
		return nil, err
	}
	return nodejs.CopyBytes(result)
}
func (m *Module) GzipSync(input any, options ...any) ([]byte, error) {
	return m.transform("gzipSync", input, options)
}
func (m *Module) GunzipSync(input any, options ...any) ([]byte, error) {
	return m.transform("gunzipSync", input, options)
}
func (m *Module) DeflateSync(input any, options ...any) ([]byte, error) {
	return m.transform("deflateSync", input, options)
}
func (m *Module) InflateSync(input any, options ...any) ([]byte, error) {
	return m.transform("inflateSync", input, options)
}
func (m *Module) DeflateRawSync(input any, options ...any) ([]byte, error) {
	return m.transform("deflateRawSync", input, options)
}
func (m *Module) InflateRawSync(input any, options ...any) ([]byte, error) {
	return m.transform("inflateRawSync", input, options)
}
func (m *Module) BrotliCompressSync(input any, options ...any) ([]byte, error) {
	return m.transform("brotliCompressSync", input, options)
}
func (m *Module) BrotliDecompressSync(input any, options ...any) ([]byte, error) {
	return m.transform("brotliDecompressSync", input, options)
}
func (m *Module) UnzipSync(input any, options ...any) ([]byte, error) {
	return m.transform("unzipSync", input, options)
}
