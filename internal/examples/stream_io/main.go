package main

import (
	"crypto/sha256"
	"io"
	"strings"
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/stream"
)

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	if err := export.SetMethod("read", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		module, err := stream.Load(ci.Env)
		if err != nil {
			return nil, err
		}
		reader := io.MultiReader(strings.NewReader("hello "), strings.NewReader("from Go\n"))
		return module.FromReader(reader, stream.IOOptions{ChunkSize: 4})
	}); err != nil {
		panic(err)
	}
	if err := export.SetMethod("digest", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		if len(ci.Args) != 1 {
			return nil, io.ErrUnexpectedEOF
		}
		// hash.Hash is an io.Writer. Node pipeline drives Go writes off-thread.
		digest := sha256.New()
		readable := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		promise, err := readable.CopyTo(digest)
		if err != nil {
			return nil, err
		}
		return promise.Then(func(call *napi.CallbackInfo) (napi.ValueType, error) {
			// Pipeline completed all writes before this callback observes the hash.
			return napi.CopyBuffer(call.Env, digest.Sum(nil))
		}, nil)
	}); err != nil {
		panic(err)
	}
	if err := export.SetMethod("echo", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		module, err := stream.Load(ci.Env)
		if err != nil {
			return nil, err
		}
		reader, writer := io.Pipe()
		value, err := module.FromDuplex(reader, writer, stream.IOOptions{CloseOnDestroy: true, ChunkSize: 4})
		if err != nil {
			reader.Close()
			writer.Close()
		}
		return value, err
	}); err != nil {
		panic(err)
	}
}

func main() {}
