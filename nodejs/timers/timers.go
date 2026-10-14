// Package timers schedules callbacks on the Node event loop. Scoped timer
// handles and module wrappers must not be cached in Go across callback scopes.
package timers

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }
type Promises struct{ *nodejs.Module }
type Timeout struct{ *napi.Object }
type Immediate struct{ *napi.Object }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:timers")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func LoadPromises(env napi.EnvType) (*Promises, error) {
	m, err := nodejs.Load(env, "node:timers/promises")
	if err != nil {
		return nil, err
	}
	return &Promises{m}, nil
}
func (m *Module) schedule(method string, callback any, delay float64, args []any) (*Timeout, error) {
	fn, err := nodejs.CallbackOf(m.Env(), callback)
	if err != nil {
		return nil, err
	}
	value, err := m.Call(method, append([]any{fn, delay}, args...)...)
	if err != nil {
		return nil, err
	}
	object, err := napi.ObjectOf(m.Env(), value)
	if err != nil {
		return nil, err
	}
	return &Timeout{object}, nil
}

// delay is a JS Number in milliseconds. Node applies its native clamping rules.
func (m *Module) SetTimeout(callback any, delay float64, args ...any) (*Timeout, error) {
	return m.schedule("setTimeout", callback, delay, args)
}
func (m *Module) SetInterval(callback any, delay float64, args ...any) (*Timeout, error) {
	return m.schedule("setInterval", callback, delay, args)
}
func (m *Module) SetImmediate(callback any, args ...any) (*Immediate, error) {
	fn, err := nodejs.CallbackOf(m.Env(), callback)
	if err != nil {
		return nil, err
	}
	value, err := m.Call("setImmediate", append([]any{fn}, args...)...)
	if err != nil {
		return nil, err
	}
	object, err := napi.ObjectOf(m.Env(), value)
	if err != nil {
		return nil, err
	}
	return &Immediate{object}, nil
}
func (m *Module) ClearTimeout(timer napi.ValueType) error {
	_, err := m.Call("clearTimeout", timer)
	return err
}
func (m *Module) ClearInterval(timer napi.ValueType) error {
	_, err := m.Call("clearInterval", timer)
	return err
}
func (m *Module) ClearImmediate(timer napi.ValueType) error {
	_, err := m.Call("clearImmediate", timer)
	return err
}
func (t *Timeout) Ref() error   { _, err := t.CallMethod("ref"); return err }
func (t *Timeout) Unref() error { _, err := t.CallMethod("unref"); return err }
func (t *Timeout) HasRef() (bool, error) {
	value, err := t.CallMethod("hasRef")
	if err != nil {
		return false, err
	}
	return napi.Decode[bool](value)
}
func (t *Timeout) Refresh() error { _, err := t.CallMethod("refresh"); return err }
func (t *Immediate) Ref() error   { _, err := t.CallMethod("ref"); return err }
func (t *Immediate) Unref() error { _, err := t.CallMethod("unref"); return err }
func (t *Immediate) HasRef() (bool, error) {
	value, err := t.CallMethod("hasRef")
	if err != nil {
		return false, err
	}
	return napi.Decode[bool](value)
}
func (p *Promises) SetTimeout(delay float64, value any, options ...any) (*nodejs.Promise, error) {
	return p.CallPromise("setTimeout", append([]any{delay, value}, options...)...)
}
func (p *Promises) SetImmediate(value any, options ...any) (*nodejs.Promise, error) {
	return p.CallPromise("setImmediate", append([]any{value}, options...)...)
}
