package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/stream"
	"sync"
	"sync/atomic"
	"time"
)

var streamIOClosed atomic.Int32

type ioProbe struct {
	mu                                            sync.Mutex
	data                                          bytes.Buffer
	reader                                        io.Reader
	mode                                          string
	delay                                         time.Duration
	calls, writes, flushes, closes, busy, maxBusy atomic.Int32
	gate                                          chan struct{}
	once                                          sync.Once
	readCloser                                    io.Closer
}

type ioHalfProbe struct {
	reader             *io.PipeReader
	writer             *io.PipeWriter
	closed, halfClosed atomic.Int32
}

func (p *ioHalfProbe) Read(data []byte) (int, error)  { return p.reader.Read(data) }
func (p *ioHalfProbe) Write(data []byte) (int, error) { return p.writer.Write(data) }
func (p *ioHalfProbe) CloseWrite() error              { p.halfClosed.Add(1); return p.writer.Close() }
func (p *ioHalfProbe) Close() error {
	p.closed.Add(1)
	return errors.Join(p.reader.Close(), p.writer.Close())
}

func (p *ioProbe) enter() {
	current := p.busy.Add(1)
	for old := p.maxBusy.Load(); current > old; old = p.maxBusy.Load() {
		if p.maxBusy.CompareAndSwap(old, current) {
			break
		}
	}
}
func (p *ioProbe) Read(data []byte) (int, error) {
	p.enter()
	defer p.busy.Add(-1)
	p.calls.Add(1)
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	switch p.mode {
	case "panic":
		panic("reader panic")
	case "emptyReads":
		return 0, nil
	case "invalid":
		return len(data) + 1, nil
	case "bothEOF":
		return copy(data, []byte("last")), io.EOF
	case "error":
		return copy(data, []byte("last")), errors.New("reader failure")
	}
	return p.reader.Read(data)
}
func (p *ioProbe) Write(data []byte) (int, error) {
	p.enter()
	defer p.busy.Add(-1)
	p.writes.Add(1)
	if p.delay > 0 {
		time.Sleep(p.delay)
	}
	if p.mode == "blocked" {
		<-p.gate
	}
	switch p.mode {
	case "panic":
		panic("writer panic")
	case "invalid":
		return len(data) + 1, nil
	case "error":
		return 0, errors.New("writer failure")
	case "short":
		if len(data) > 0 {
			return len(data) - 1, nil
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.data.Write(data)
}
func (p *ioProbe) Flush() error {
	p.flushes.Add(1)
	if p.mode == "flushError" {
		return errors.New("flush failure")
	}
	return nil
}
func (p *ioProbe) Close() error {
	p.closes.Add(1)
	streamIOClosed.Add(1)
	if p.mode == "blocked" {
		p.once.Do(func() { close(p.gate) })
	}
	if p.readCloser != nil {
		return p.readCloser.Close()
	}
	if p.mode == "closeError" {
		return errors.New("close failure")
	}
	return nil
}
func (p *ioProbe) stats(env napi.EnvType) napi.ValueType {
	p.mu.Lock()
	data := append([]byte{}, p.data.Bytes()...)
	p.mu.Unlock()
	return must(napi.ObjectOf(env, map[string]any{"reads": int(p.calls.Load()), "writes": int(p.writes.Load()), "flushes": int(p.flushes.Load()), "closes": int(p.closes.Load()), "busy": int(p.busy.Load()), "maxBusy": int(p.maxBusy.Load()), "data": data}))
}
func ioControl(env napi.EnvType, s *stream.Stream, p *ioProbe, cancel context.CancelFunc) *napi.Object {
	control := must(napi.ObjectOf(env, map[string]any{"stream": s}))
	if err := control.SetMethod("stats", func(ci *napi.CallbackInfo) (napi.ValueType, error) { return p.stats(ci.Env), nil }); err != nil {
		panic(err)
	}
	if err := control.SetMethod("cancel", func(ci *napi.CallbackInfo) (napi.ValueType, error) { cancel(); return nil, nil }); err != nil {
		panic(err)
	}
	if err := control.SetMethod("release", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		p.once.Do(func() { close(p.gate) })
		return nil, nil
	}); err != nil {
		panic(err)
	}
	return control
}
func registerStreamIO(env napi.EnvType, export *napi.Object) {
	bind := func(name string, callback napi.Callback) {
		if err := export.SetMethod(name, callback); err != nil {
			panic(err)
		}
	}
	bind("ioReadable", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		mode := must(napi.ArgAs[string](ci, 0))
		data := must(napi.ArgAs[string](ci, 1))
		owned := must(napi.ArgAs[bool](ci, 2))
		ctx, cancel := context.WithCancel(context.Background())
		if mode == "deadline" {
			cancel()
			ctx, cancel = context.WithTimeout(context.Background(), 30*time.Millisecond)
		}
		p := &ioProbe{mode: mode, reader: bytes.NewReader([]byte(data)), gate: make(chan struct{})}
		if mode == "slow" {
			p.delay = 20 * time.Millisecond
		}
		if mode == "pipe" || mode == "deadline" {
			reader, writer := io.Pipe()
			p.reader = reader
			p.readCloser = reader
			defer func() {
				if ctx.Err() != nil {
					writer.Close()
				}
			}()
		}
		m := must(stream.Load(ci.Env))
		s, err := m.FromReader(p, stream.IOOptions{Context: ctx, ChunkSize: 4, HighWaterMark: 4, CloseOnDestroy: owned})
		if err != nil {
			cancel()
			return nil, err
		}
		return ioControl(ci.Env, s, p, cancel), nil
	})
	bind("ioWritable", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		mode := must(napi.ArgAs[string](ci, 0))
		owned := must(napi.ArgAs[bool](ci, 1))
		ctx, cancel := context.WithCancel(context.Background())
		p := &ioProbe{mode: mode, gate: make(chan struct{})}
		if mode == "slow" {
			p.delay = 20 * time.Millisecond
		}
		m := must(stream.Load(ci.Env))
		s, err := m.FromWriter(p, stream.IOOptions{Context: ctx, HighWaterMark: 1, CloseOnDestroy: owned})
		if err != nil {
			cancel()
			return nil, err
		}
		return ioControl(ci.Env, s, p, cancel), nil
	})
	bind("ioDuplex", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		reader, writer := io.Pipe()
		m := must(stream.Load(ci.Env))
		return m.FromDuplex(reader, writer, stream.IOOptions{CloseOnDestroy: true, ChunkSize: 3, HighWaterMark: 3})
	})
	bind("ioReadWriter", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		endpoint, peer := net.Pipe()
		go func() { defer peer.Close(); _, _ = io.Copy(peer, peer) }()
		m := must(stream.Load(ci.Env))
		return m.FromReadWriter(endpoint, stream.IOOptions{CloseOnDestroy: true, ChunkSize: 3})
	})
	bind("ioHalfClose", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		reader, writer := io.Pipe()
		probe := &ioHalfProbe{reader: reader, writer: writer}
		module := must(stream.Load(ci.Env))
		value := must(module.FromReadWriter(probe, stream.IOOptions{CloseOnDestroy: true, ChunkSize: 3}))
		control := must(napi.ObjectOf(ci.Env, map[string]any{"stream": value}))
		if err := control.SetMethod("stats", func(call *napi.CallbackInfo) (napi.ValueType, error) {
			return napi.ObjectOf(call.Env, map[string]any{"closed": int(probe.closed.Load()), "halfClosed": int(probe.halfClosed.Load())})
		}); err != nil {
			return nil, err
		}
		return control, nil
	})
	bind("ioCopyTo", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		p := &ioProbe{mode: "bytes", gate: make(chan struct{})}
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		promise, err := s.CopyTo(p)
		if err != nil {
			return nil, err
		}
		return promise.Then(func(call *napi.CallbackInfo) (napi.ValueType, error) {
			p.mu.Lock()
			data := append([]byte{}, p.data.Bytes()...)
			p.mu.Unlock()
			return napi.CopyBuffer(call.Env, data)
		}, nil)
	})
	bind("ioCopyFrom", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		data, err := nodejs.CopyBytes(ci.Args[1])
		if err != nil {
			return nil, err
		}
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		return s.CopyFrom(bytes.NewReader(data), stream.IOOptions{ChunkSize: 3})
	})
	bind("ioCopyError", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		p := &ioProbe{mode: "error", gate: make(chan struct{})}
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		return s.CopyTo(p)
	})
	bind("ioInvalid", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(stream.Load(ci.Env))
		mode := must(napi.ArgAs[string](ci, 0))
		switch mode {
		case "reader":
			return m.FromReader((*bytes.Reader)(nil))
		case "writer":
			return m.FromWriter((*bytes.Buffer)(nil))
		case "duplex":
			return m.FromDuplex(nil, &bytes.Buffer{})
		case "rw":
			return m.FromReadWriter((*bytes.Buffer)(nil))
		case "options":
			return m.FromReader(bytes.NewReader(nil), stream.IOOptions{ChunkSize: -1})
		case "twoOptions":
			return m.FromReader(bytes.NewReader(nil), stream.IOOptions{}, stream.IOOptions{})
		case "context":
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return m.FromReader(bytes.NewReader(nil), stream.IOOptions{Context: ctx})
		}
		return nil, errors.New("unknown validation")
	})
	bind("ioClosedCount", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		return napi.ValueOf(ci.Env, int(streamIOClosed.Load()))
	})
}
