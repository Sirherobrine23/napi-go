// Package path delegates path operations to Node, including explicit POSIX/Windows variants.
package path

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }
type Parts struct {
	Root string `napi:"root"`
	Dir  string `napi:"dir"`
	Base string `napi:"base"`
	Ext  string `napi:"ext"`
	Name string `napi:"name"`
}

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:path")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) variant(name string) (*Module, error) {
	value, err := m.Get(name)
	if err != nil {
		return nil, err
	}
	child, err := nodejs.Wrap(value, m.ID+"."+name)
	if err != nil {
		return nil, err
	}
	return &Module{child}, nil
}
func (m *Module) Posix() (*Module, error) { return m.variant("posix") }
func (m *Module) Win32() (*Module, error) { return m.variant("win32") }
func (m *Module) Join(paths ...string) (string, error) {
	return nodejs.CallAs[string](m.Module, "join", nodejs.Strings(paths)...)
}
func (m *Module) Resolve(paths ...string) (string, error) {
	return nodejs.CallAs[string](m.Module, "resolve", nodejs.Strings(paths)...)
}
func (m *Module) Basename(path string, suffix ...string) (string, error) {
	return nodejs.CallAs[string](m.Module, "basename", append([]any{path}, nodejs.Strings(suffix)...)...)
}
func (m *Module) Dirname(path string) (string, error) {
	return nodejs.CallAs[string](m.Module, "dirname", path)
}
func (m *Module) Extname(path string) (string, error) {
	return nodejs.CallAs[string](m.Module, "extname", path)
}
func (m *Module) Normalize(path string) (string, error) {
	return nodejs.CallAs[string](m.Module, "normalize", path)
}
func (m *Module) Relative(from, to string) (string, error) {
	return nodejs.CallAs[string](m.Module, "relative", from, to)
}
func (m *Module) IsAbsolute(path string) (bool, error) {
	return nodejs.CallAs[bool](m.Module, "isAbsolute", path)
}
func (m *Module) Parse(path string) (Parts, error) {
	return nodejs.CallAs[Parts](m.Module, "parse", path)
}
func (m *Module) Format(parts Parts) (string, error) {
	return nodejs.CallAs[string](m.Module, "format", parts)
}
func (m *Module) Separator() (string, error) { return napi.GetAs[string](m.Object, "sep") }
func (m *Module) Delimiter() (string, error) { return napi.GetAs[string](m.Object, "delimiter") }
