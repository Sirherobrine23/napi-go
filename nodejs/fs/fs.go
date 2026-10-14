// Package fs delegates filesystem operations to Node. Sync methods run on and
// block the Node thread. LoadPromises provides async operations returned to JS.
package fs

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }
type Promises struct{ *nodejs.Module }
type Stats struct{ *napi.Object }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:fs")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func LoadPromises(env napi.EnvType) (*Promises, error) {
	m, err := nodejs.Load(env, "node:fs/promises")
	if err != nil {
		return nil, err
	}
	return &Promises{m}, nil
}

// ReadFileSync returns a scoped Buffer (or a string when encoding is supplied).
func (m *Module) ReadFileSync(path any, options ...any) (napi.ValueType, error) {
	return m.Call("readFileSync", append([]any{path}, options...)...)
}
func (m *Module) ReadBytesSync(path any) ([]byte, error) {
	value, err := m.ReadFileSync(path)
	if err != nil {
		return nil, err
	}
	return nodejs.CopyBytes(value)
}
func (m *Module) ReadTextSync(path any, encoding ...string) (string, error) {
	codec := "utf8"
	if len(encoding) > 0 {
		codec = encoding[0]
	}
	return nodejs.CallAs[string](m.Module, "readFileSync", path, codec)
}
func (m *Module) WriteFileSync(path any, data any, options ...any) error {
	value, err := nodejs.Binary(m.Env(), data)
	if err != nil {
		return err
	}
	_, err = m.Call("writeFileSync", append([]any{path, value}, options...)...)
	return err
}
func (m *Module) AppendFileSync(path any, data any, options ...any) error {
	value, err := nodejs.Binary(m.Env(), data)
	if err != nil {
		return err
	}
	_, err = m.Call("appendFileSync", append([]any{path, value}, options...)...)
	return err
}
func (m *Module) ExistsSync(path any) (bool, error) {
	return nodejs.CallAs[bool](m.Module, "existsSync", path)
}
func (m *Module) ReaddirSync(path any) ([]string, error) {
	return nodejs.CallAs[[]string](m.Module, "readdirSync", path)
}
func (m *Module) StatSync(path any, options ...any) (*Stats, error) {
	value, err := m.Call("statSync", append([]any{path}, options...)...)
	if err != nil {
		return nil, err
	}
	kind, err := value.Type()
	if err != nil {
		return nil, err
	}
	if kind == napi.TypeUndefined {
		// Native throwIfNoEntry:false reports an absent entry as undefined.
		return nil, nil
	}
	object, err := napi.ObjectOf(m.Env(), value)
	if err != nil {
		return nil, err
	}
	return &Stats{object}, nil
}
func (m *Module) MkdirSync(path any, options ...any) (napi.ValueType, error) {
	return m.Call("mkdirSync", append([]any{path}, options...)...)
}
func (m *Module) MkdtempSync(prefix string) (string, error) {
	return nodejs.CallAs[string](m.Module, "mkdtempSync", prefix)
}
func (m *Module) RenameSync(from, to any) error { _, err := m.Call("renameSync", from, to); return err }
func (m *Module) UnlinkSync(path any) error     { _, err := m.Call("unlinkSync", path); return err }
func (m *Module) RmSync(path any, options ...any) error {
	_, err := m.Call("rmSync", append([]any{path}, options...)...)
	return err
}
func (s *Stats) IsFile() (bool, error) {
	value, err := s.CallMethod("isFile")
	if err != nil {
		return false, err
	}
	return napi.Decode[bool](value)
}
func (s *Stats) IsDirectory() (bool, error) {
	value, err := s.CallMethod("isDirectory")
	if err != nil {
		return false, err
	}
	return napi.Decode[bool](value)
}
func (s *Stats) IsSymbolicLink() (bool, error) {
	value, err := s.CallMethod("isSymbolicLink")
	if err != nil {
		return false, err
	}
	return napi.Decode[bool](value)
}
func (s *Stats) Size() (int64, error) { return napi.GetAs[int64](s.Object, "size") }
func (s *Stats) Mode() (int, error)   { return napi.GetAs[int](s.Object, "mode") }
func (p *Promises) ReadFile(path any, options ...any) (*nodejs.Promise, error) {
	return p.CallPromise("readFile", append([]any{path}, options...)...)
}
func (p *Promises) WriteFile(path any, data any, options ...any) (*nodejs.Promise, error) {
	value, err := nodejs.Binary(p.Env(), data)
	if err != nil {
		return nil, err
	}
	return p.CallPromise("writeFile", append([]any{path, value}, options...)...)
}
func (p *Promises) Stat(path any, options ...any) (*nodejs.Promise, error) {
	return p.CallPromise("stat", append([]any{path}, options...)...)
}
func (p *Promises) Mkdir(path any, options ...any) (*nodejs.Promise, error) {
	return p.CallPromise("mkdir", append([]any{path}, options...)...)
}
func (p *Promises) Readdir(path any, options ...any) (*nodejs.Promise, error) {
	return p.CallPromise("readdir", append([]any{path}, options...)...)
}
func (p *Promises) Rm(path any, options ...any) (*nodejs.Promise, error) {
	return p.CallPromise("rm", append([]any{path}, options...)...)
}
