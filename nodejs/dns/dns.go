// Package dns provides Promise-based Node DNS operations. Lookup uses the host
// resolver, while Resolve methods use DNS queries. Tests only look up localhost.
package dns

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:dns/promises")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) Lookup(host string, options ...any) (*nodejs.Promise, error) {
	return m.CallPromise("lookup", append([]any{host}, options...)...)
}
func (m *Module) Resolve(host string, rrtype ...string) (*nodejs.Promise, error) {
	return m.CallPromise("resolve", append([]any{host}, nodejs.Strings(rrtype)...)...)
}
func (m *Module) Resolve4(host string, options ...any) (*nodejs.Promise, error) {
	return m.CallPromise("resolve4", append([]any{host}, options...)...)
}
func (m *Module) Resolve6(host string, options ...any) (*nodejs.Promise, error) {
	return m.CallPromise("resolve6", append([]any{host}, options...)...)
}
func (m *Module) Reverse(address string) (*nodejs.Promise, error) {
	return m.CallPromise("reverse", address)
}
