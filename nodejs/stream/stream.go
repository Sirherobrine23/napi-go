// Package stream wraps Node streams and their event-driven backpressure contract.
package stream

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
	event "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/EventEmitter"
)

type Module struct{ *nodejs.Module }
type Stream struct{ *napi.Object }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:stream")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) PassThrough(options ...any) (*Stream, error) {
	value, err := m.New("PassThrough", options...)
	if err != nil {
		return nil, err
	}
	return &Stream{value}, nil
}
func (m *Module) ReadableFrom(iterable any, options ...any) (*Stream, error) {
	readable, err := m.Get("Readable")
	if err != nil {
		return nil, err
	}
	value, err := napi.ToObject(readable).CallMethod("from", append([]any{iterable}, options...)...)
	if err != nil {
		return nil, err
	}
	object, err := napi.ObjectOf(m.Env(), value)
	if err != nil {
		return nil, err
	}
	return &Stream{object}, nil
}

// Write returns false when the producer must wait for the drain event. False is
// not a failure, and no synchronous Go wait is introduced on the Node thread.
func (s *Stream) Write(data any, options ...any) (bool, error) {
	value, err := nodejs.Binary(s.Env(), data)
	if err != nil {
		return false, err
	}
	result, err := s.CallMethod("write", append([]any{value}, options...)...)
	if err != nil {
		return false, err
	}
	return napi.Decode[bool](result)
}
func (s *Stream) End(args ...any) error {
	if len(args) > 0 {
		value, err := nodejs.Binary(s.Env(), args[0])
		if err != nil {
			return err
		}
		args = append([]any{value}, args[1:]...)
	}
	_, err := s.CallMethod("end", args...)
	return err
}
func (s *Stream) Pipe(destination napi.ValueType, options ...any) (napi.ValueType, error) {
	return s.CallMethod("pipe", append([]any{destination}, options...)...)
}
func (s *Stream) Read(size ...int) (napi.ValueType, error) {
	args := make([]any, len(size))
	for i, v := range size {
		args[i] = v
	}
	return s.CallMethod("read", args...)
}
func (s *Stream) Destroy(args ...any) error { _, err := s.CallMethod("destroy", args...); return err }
func (s *Stream) On(name string, callback any) (*napi.Function, error) {
	fn, err := nodejs.CallbackOf(s.Env(), callback)
	if err != nil {
		return nil, err
	}
	_, err = s.CallMethod("on", name, fn)
	if err != nil {
		return nil, err
	}
	return fn, nil
}
func (s *Stream) Once(name string, callback any) (*napi.Function, error) {
	fn, err := nodejs.CallbackOf(s.Env(), callback)
	if err != nil {
		return nil, err
	}
	_, err = s.CallMethod("once", name, fn)
	if err != nil {
		return nil, err
	}
	return fn, nil
}
func (s *Stream) Events() *event.EventEmitter { return event.ToEventEmitter(s.Object) }
