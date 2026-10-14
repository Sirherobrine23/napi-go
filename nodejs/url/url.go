// Package url wraps WHATWG URL and URLSearchParams as scoped JavaScript objects.
package url

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }
type URL struct{ *napi.Object }
type SearchParams struct{ *napi.Object }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:url")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) NewURL(input string, base ...string) (*URL, error) {
	value, err := m.New("URL", append([]any{input}, nodejs.Strings(base)...)...)
	if err != nil {
		return nil, err
	}
	return &URL{value}, nil
}
func (m *Module) NewSearchParams(input ...any) (*SearchParams, error) {
	value, err := m.New("URLSearchParams", input...)
	if err != nil {
		return nil, err
	}
	return &SearchParams{value}, nil
}
func (u *URL) SearchParams() (*SearchParams, error) {
	value, err := u.Get("searchParams")
	if err != nil {
		return nil, err
	}
	object, err := napi.ObjectOf(u.Env(), value)
	if err != nil {
		return nil, err
	}
	return &SearchParams{object}, nil
}
func (p *SearchParams) Get(name string) (string, bool, error) {
	value, err := p.CallMethod("get", name)
	if err != nil {
		return "", false, err
	}
	kind, err := value.Type()
	if err != nil {
		return "", false, err
	}
	if kind == napi.TypeNull {
		return "", false, nil
	}
	text, err := napi.Decode[string](value)
	return text, true, err
}
func (p *SearchParams) GetAll(name string) ([]string, error) {
	value, err := p.CallMethod("getAll", name)
	if err != nil {
		return nil, err
	}
	return napi.Decode[[]string](value)
}
func (p *SearchParams) Has(name string) (bool, error) {
	value, err := p.CallMethod("has", name)
	if err != nil {
		return false, err
	}
	return napi.Decode[bool](value)
}
func (p *SearchParams) Append(name, value string) error {
	_, err := p.CallMethod("append", name, value)
	return err
}
func (p *SearchParams) Set(name, value string) error {
	_, err := p.CallMethod("set", name, value)
	return err
}
func (p *SearchParams) Delete(name string) error { _, err := p.CallMethod("delete", name); return err }
func (p *SearchParams) Sort() error              { _, err := p.CallMethod("sort"); return err }
func (p *SearchParams) String() (string, error) {
	value, err := p.CallMethod("toString")
	if err != nil {
		return "", err
	}
	return napi.Decode[string](value)
}
func (m *Module) FileURLToPath(value any, options ...any) (string, error) {
	return nodejs.CallAs[string](m.Module, "fileURLToPath", append([]any{value}, options...)...)
}
func (m *Module) PathToFileURL(path string, options ...any) (*URL, error) {
	value, err := m.Call("pathToFileURL", append([]any{path}, options...)...)
	if err != nil {
		return nil, err
	}
	object, err := napi.ObjectOf(m.Env(), value)
	if err != nil {
		return nil, err
	}
	return &URL{object}, nil
}
func (m *Module) DomainToASCII(domain string) (string, error) {
	return nodejs.CallAs[string](m.Module, "domainToASCII", domain)
}
func (m *Module) DomainToUnicode(domain string) (string, error) {
	return nodejs.CallAs[string](m.Module, "domainToUnicode", domain)
}
func (u *URL) Href() (string, error)     { return napi.GetAs[string](u.Object, "href") }
func (u *URL) Origin() (string, error)   { return napi.GetAs[string](u.Object, "origin") }
func (u *URL) Protocol() (string, error) { return napi.GetAs[string](u.Object, "protocol") }
func (u *URL) Hostname() (string, error) { return napi.GetAs[string](u.Object, "hostname") }
func (u *URL) Port() (string, error)     { return napi.GetAs[string](u.Object, "port") }
func (u *URL) Pathname() (string, error) { return napi.GetAs[string](u.Object, "pathname") }
func (u *URL) Search() (string, error)   { return napi.GetAs[string](u.Object, "search") }
func (u *URL) Hash() (string, error)     { return napi.GetAs[string](u.Object, "hash") }
