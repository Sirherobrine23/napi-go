package main

import (
	_ "unsafe"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	_ "sirherobrine23.com.br/Sirherobrine23/napi-go/module"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/crypto"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/fs"
	nodepath "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/path"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/timers"
)

func hashFile(ci *napi.CallbackInfo) (napi.ValueType, error) {
	filename, err := napi.ArgAs[string](ci, 0)
	if err != nil {
		return nil, err
	}
	files, err := fs.LoadPromises(ci.Env)
	if err != nil {
		return nil, err
	}
	promise, err := files.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	// The Node Promise retains this callback. Load crypto in its fresh scope;
	// do not capture files or other JS handles from hashFile's earlier scope.
	return promise.Then(func(call *napi.CallbackInfo) (napi.ValueType, error) {
		module, err := crypto.Load(call.Env)
		if err != nil {
			return nil, err
		}
		hash, err := module.CreateHash("sha256")
		if err != nil {
			return nil, err
		}
		if err := hash.Update(call.Args[0]); err != nil {
			return nil, err
		}
		return hash.Digest("hex")
	}, nil)
}

//go:linkname Register sirherobrine23.com.br/Sirherobrine23/napi-go/module.Register
func Register(env napi.EnvType, export *napi.Object) {
	if err := export.SetMethod("hashFile", hashFile); err != nil {
		panic(err)
	}
	if err := export.SetMethod("join", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		parts, err := napi.ArgAs[[]string](ci, 0)
		if err != nil {
			return nil, err
		}
		module, err := nodepath.Load(ci.Env)
		if err != nil {
			return nil, err
		}
		value, err := module.Join(parts...)
		if err != nil {
			return nil, err
		}
		return napi.ValueOf(ci.Env, value)
	}); err != nil {
		panic(err)
	}
	if err := export.SetMethod("later", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		module, err := timers.LoadPromises(ci.Env)
		if err != nil {
			return nil, err
		}
		return module.SetTimeout(1, "from the Node event loop")
	}); err != nil {
		panic(err)
	}
}

func main() {}
