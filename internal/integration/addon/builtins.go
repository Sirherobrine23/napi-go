package main

import (
	"fmt"
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
	nodeevents "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/EventEmitter"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/buffer"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/crypto"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/dns"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/fs"
	nodeos "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/os"
	nodepath "sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/path"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/stream"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/timers"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/url"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/util"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/zlib"
	"strings"
)

func registerBuiltins(env napi.EnvType, export *napi.Object) {
	bind := func(name string, cb napi.Callback) {
		if err := export.SetMethod(name, cb); err != nil {
			panic(err)
		}
	}
	bind("builtinLoad", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		id, err := napi.ArgAs[string](ci, 0)
		if err != nil {
			return nil, err
		}
		return nodejs.Load(ci.Env, id)
	})
	bind("builtinCall", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		id, err := napi.ArgAs[string](ci, 0)
		if err != nil {
			return nil, err
		}
		method, err := napi.ArgAs[string](ci, 1)
		if err != nil {
			return nil, err
		}
		m, err := nodejs.Load(ci.Env, id)
		if err != nil {
			return nil, err
		}
		args := make([]any, len(ci.Args)-2)
		for i, value := range ci.Args[2:] {
			args[i] = value
		}
		return m.Call(method, args...)
	})
	bind("builtinWrap", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m, err := nodejs.Wrap(ci.Args[0], "custom")
		if err != nil {
			return nil, err
		}
		return m.Call("method", 23)
	})
	bind("builtinPath", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(nodepath.Load(ci.Env))
		flavor := must(napi.ArgAs[string](ci, 0))
		input := must(napi.ArgAs[string](ci, 1))
		if flavor == "win32" {
			m = must(m.Win32())
		} else if flavor == "posix" {
			m = must(m.Posix())
		}
		parts := must(m.Parse(input))
		return napi.ObjectOf(ci.Env, map[string]any{"parsed": parts, "formatted": must(m.Format(parts)), "basename": must(m.Basename(input)), "stem": must(m.Basename(input, parts.Ext)), "dirname": must(m.Dirname(input)), "extname": must(m.Extname(input)), "absolute": must(m.IsAbsolute(input)), "normalize": must(m.Normalize(input)), "join": must(m.Join("a", "b", "..", "c")), "resolve": must(m.Resolve(".")), "relative": must(m.Relative("a", "b")), "sep": must(m.Separator()), "delimiter": must(m.Delimiter())})
	})
	bind("builtinOS", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(nodeos.Load(ci.Env))
		cpus := must(m.CPUs())
		user := must(m.UserInfo())
		return napi.ObjectOf(ci.Env, map[string]any{"arch": must(m.Arch()), "platform": must(m.Platform()), "hostname": must(m.Hostname()), "type": must(m.OSType()), "release": must(m.Release()), "homedir": must(m.Homedir()), "tmpdir": must(m.Tmpdir()), "endianness": must(m.Endianness()), "totalmem": must(m.Totalmem()), "freemem": must(m.Freemem()), "uptime": must(m.Uptime()), "loadavg": must(m.Loadavg()), "parallel": must(m.AvailableParallelism()), "cpus": cpus, "user": user})
	})
	bind("builtinNetworkInterfaces", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(nodeos.Load(ci.Env))
		return m.NetworkInterfaces()
	})
	bind("builtinBuffer", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(buffer.Load(ci.Env))
		a := must(m.From("héllo"))
		b := must(m.From([]byte{0, 255, 1}))
		merged := must(m.Concat(a, b))
		data := must(buffer.CopyBytes(b))
		binary := must(nodejs.Binary(ci.Env, []byte{}))
		return napi.ObjectOf(ci.Env, map[string]any{"text": must(buffer.ToString(a)), "hex": must(buffer.ToString(a, "hex")), "length": must(m.ByteLength("héllo")), "buffer": merged, "bytes": data, "isBuffer": must(m.IsBuffer(b)), "isArrayBuffer": must(m.IsBuffer([]byte{1})), "alloc": must(m.Alloc(3, 7)), "empty": must(m.Alloc(0)), "emptyCopy": binary, "compare": must(m.Compare(a, a)), "type": must(b.Type()).String()})
	})
	bind("builtinBufferClone", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		data, err := nodejs.CopyBytes(ci.Args[0])
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if err := napi.ToObject(ci.Args[0]).SetValue("0", 99); err != nil {
				return nil, err
			}
		}
		return napi.CopyBuffer(ci.Env, data)
	})
	bind("builtinHash", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(crypto.Load(ci.Env))
		algorithm := must(napi.ArgAs[string](ci, 0))
		data := must(napi.ArgAs[string](ci, 1))
		hash, err := m.CreateHash(algorithm)
		if err != nil {
			return nil, err
		}
		if err := hash.Update([]byte(data)); err != nil {
			return nil, err
		}
		hex, err := hash.DigestHex()
		if err != nil {
			return nil, err
		}
		return napi.ValueOf(ci.Env, hex)
	})
	bind("builtinCrypto", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(crypto.Load(ci.Env))
		mac := must(m.CreateHmac("sha256", []byte("secret")))
		if err := mac.Update("data"); err != nil {
			return nil, err
		}
		hash := must(m.CreateHash("sha256"))
		if err := hash.Update([]byte{}); err != nil {
			return nil, err
		}
		digest := must(hash.DigestBytes())
		return napi.ObjectOf(ci.Env, map[string]any{"hmac": must(mac.DigestHex()), "emptyHash": digest, "emptyRandom": must(m.RandomBytes(0)), "random": must(m.RandomBytes(16)), "uuid": must(m.RandomUUID()), "equal": must(m.TimingSafeEqual([]byte{1, 2}, []byte{1, 2})), "different": must(m.TimingSafeEqual([]byte{1}, []byte{2})), "hashes": must(m.GetHashes())})
	})
	bind("builtinCryptoUnequal", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(crypto.Load(ci.Env))
		_, err := m.TimingSafeEqual([]byte{1}, []byte{1, 2})
		return nil, err
	})
	bind("builtinCryptoFinalized", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(crypto.Load(ci.Env))
		h := must(m.CreateHash("sha256"))
		must(h.DigestHex())
		return h.Digest()
	})
	bind("builtinURL", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(url.Load(ci.Env))
		input := must(napi.ArgAs[string](ci, 0))
		u, err := m.NewURL(input, "https://example.com/base/")
		if err != nil {
			return nil, err
		}
		params := must(u.SearchParams())
		missing, exists, err := params.Get("missing")
		if err != nil {
			return nil, err
		}
		if err := params.Append("q", "two words"); err != nil {
			return nil, err
		}
		if err := params.Set("empty", ""); err != nil {
			return nil, err
		}
		empty, present, err := params.Get("empty")
		if err != nil {
			return nil, err
		}
		all := must(params.GetAll("q"))
		if err := params.Delete("empty"); err != nil {
			return nil, err
		}
		if err := params.Sort(); err != nil {
			return nil, err
		}
		own := must(m.NewSearchParams([][]string{{"a", "1"}, {"a", "2"}}))
		return napi.ObjectOf(ci.Env, map[string]any{"value": u, "href": must(u.Href()), "origin": must(u.Origin()), "protocol": must(u.Protocol()), "hostname": must(u.Hostname()), "port": must(u.Port()), "pathname": must(u.Pathname()), "search": must(u.Search()), "hash": must(u.Hash()), "missing": missing, "exists": exists, "empty": empty, "present": present, "hasQ": must(params.Has("q")), "all": all, "query": must(params.String()), "pairs": must(own.String()), "ascii": must(m.DomainToASCII("münich.example")), "unicode": must(m.DomainToUnicode("xn--mnich-kva.example"))})
	})
	bind("builtinFileURL", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(url.Load(ci.Env))
		path := must(napi.ArgAs[string](ci, 0))
		u := must(m.PathToFileURL(path))
		return napi.ObjectOf(ci.Env, map[string]any{"href": must(u.Href()), "path": must(m.FileURLToPath(u))})
	})
	bind("builtinUtil", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(util.Load(ci.Env))
		return napi.ObjectOf(ci.Env, map[string]any{"formatted": must(m.Format("value=%d %s", 23, "go")), "inspected": must(m.Inspect(map[string]any{"answer": 42}, map[string]any{"compact": true})), "equal": must(m.IsDeepStrictEqual(map[string]int{"a": 1}, map[string]int{"a": 1})), "unequal": must(m.IsDeepStrictEqual([]int{1}, []int{2}))})
	})
	bind("builtinPromisify", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(util.Load(ci.Env))
		return m.Promisify(ci.Args[0])
	})
	bind("builtinCallbackify", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(util.Load(ci.Env))
		return m.Callbackify(ci.Args[0])
	})
	bind("builtinCompress", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(zlib.Load(ci.Env))
		method := must(napi.ArgAs[string](ci, 0))
		input := ci.Args[1]
		var data []byte
		var err error
		switch method {
		case "gzip":
			data, err = m.GzipSync(input)
		case "gunzip":
			data, err = m.GunzipSync(input)
		case "deflate":
			data, err = m.DeflateSync(input)
		case "inflate":
			data, err = m.InflateSync(input)
		case "deflateRaw":
			data, err = m.DeflateRawSync(input)
		case "inflateRaw":
			data, err = m.InflateRawSync(input)
		case "brotli":
			data, err = m.BrotliCompressSync(input)
		case "unbrotli":
			data, err = m.BrotliDecompressSync(input)
		case "unzip":
			data, err = m.UnzipSync(input)
		default:
			return nil, fmt.Errorf("unknown compression")
		}
		if err != nil {
			return nil, err
		}
		return napi.CopyBuffer(ci.Env, data)
	})
	bind("builtinFS", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(fs.Load(ci.Env))
		dir := must(napi.ArgAs[string](ci, 0))
		filename := dir + "/data.bin"
		if err := m.WriteFileSync(filename, []byte{0, 255, 3}); err != nil {
			return nil, err
		}
		if err := m.AppendFileSync(filename, []byte{4}); err != nil {
			return nil, err
		}
		stat := must(m.StatSync(filename))
		size := must(stat.Size())
		data := must(m.ReadBytesSync(filename))
		if err := m.WriteFileSync(dir+"/text.txt", "héllo"); err != nil {
			return nil, err
		}
		text := must(m.ReadTextSync(dir + "/text.txt"))
		entries := must(m.ReaddirSync(dir))
		tmp := must(m.MkdtempSync(dir + "/temp-"))
		if err := m.RmSync(tmp, map[string]any{"recursive": true}); err != nil {
			return nil, err
		}
		child := dir + "/child"
		must(m.MkdirSync(child, map[string]any{"recursive": true}))
		childStat := must(m.StatSync(child))
		if err := m.RmSync(child, map[string]any{"recursive": true}); err != nil {
			return nil, err
		}
		if err := m.RenameSync(filename, dir+"/moved.bin"); err != nil {
			return nil, err
		}
		if err := m.UnlinkSync(dir + "/moved.bin"); err != nil {
			return nil, err
		}
		return napi.ObjectOf(ci.Env, map[string]any{"data": data, "text": text, "size": float64(size), "file": must(stat.IsFile()), "directory": must(childStat.IsDirectory()), "symlink": must(stat.IsSymbolicLink()), "mode": must(stat.Mode()), "entries": entries, "exists": must(m.ExistsSync(filename)), "raw": must(m.ReadFileSync(dir + "/text.txt"))})
	})
	bind("builtinFSRead", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(fs.Load(ci.Env))
		return m.ReadFileSync(ci.Args[0])
	})
	bind("builtinFSStatMaybe", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(fs.Load(ci.Env))
		stat, err := m.StatSync(ci.Args[0], map[string]any{"throwIfNoEntry": false})
		if err != nil {
			return nil, err
		}
		if stat == nil {
			return ci.Env.Undefined()
		}
		return stat, nil
	})
	bind("builtinFSAsync", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(fs.LoadPromises(ci.Env))
		op := must(napi.ArgAs[string](ci, 0))
		path := ci.Args[1]
		switch op {
		case "read":
			return m.ReadFile(path)
		case "text":
			return m.ReadFile(path, "utf8")
		case "write":
			return m.WriteFile(path, []byte{1, 0, 255})
		case "stat":
			return m.Stat(path)
		case "mkdir":
			return m.Mkdir(path, map[string]any{"recursive": true})
		case "readdir":
			return m.Readdir(path)
		case "rm":
			return m.Rm(path, map[string]any{"recursive": true})
		default:
			return nil, fmt.Errorf("unknown fs operation")
		}
	})
	bind("builtinPromise", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		p, err := nodejs.ToPromise(ci.Args[0])
		if err != nil {
			return nil, err
		}
		mode := must(napi.ArgAs[string](ci, 1))
		switch mode {
		case "upper":
			return p.Then(func(call *napi.CallbackInfo) (napi.ValueType, error) {
				text, err := napi.ArgAs[string](call, 0)
				if err != nil {
					return nil, err
				}
				return napi.ValueOf(call.Env, strings.ToUpper(text))
			}, nil)
		case "fail":
			return p.Then(func(call *napi.CallbackInfo) (napi.ValueType, error) {
				return nil, fmt.Errorf("Go promise handler failure")
			}, nil)
		case "catch":
			return p.Catch(func(call *napi.CallbackInfo) (napi.ValueType, error) {
				message, err := napi.GetAs[string](napi.ToObject(call.Args[0]), "message")
				if err != nil {
					return nil, err
				}
				return napi.ValueOf(call.Env, "caught:"+message)
			})
		case "finally":
			return p.Finally(ci.Args[2])
		case "identity":
			return p.Then(nil, nil)
		case "adopt":
			return p.Then(func(call *napi.CallbackInfo) (napi.ValueType, error) {
				m := must(timers.LoadPromises(call.Env))
				return m.SetTimeout(1, "adopted")
			}, nil)
		case "invalid":
			return p.Then(23, nil)
		default:
			return nil, fmt.Errorf("unknown promise case")
		}
	})
	bind("builtinPromiseChain", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		p, err := nodejs.ToPromise(ci.Args[0])
		if err != nil {
			return nil, err
		}
		method := must(napi.ArgAs[string](ci, 1))
		var handlers [2]any
		for index := range handlers {
			if len(ci.Args) > index+2 {
				handlers[index] = ci.Args[index+2]
			}
		}
		switch method {
		case "then":
			return p.Then(handlers[0], handlers[1])
		case "catch":
			return p.Catch(handlers[0])
		case "finally":
			return p.Finally(handlers[0])
		default:
			return nil, fmt.Errorf("unknown promise method %q", method)
		}
	})
	bind("builtinTimer", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(timers.Load(ci.Env))
		mode := must(napi.ArgAs[string](ci, 1))
		switch mode {
		case "immediate":
			return m.SetImmediate(ci.Args[0], "immediate")
		case "interval":
			return m.SetInterval(ci.Args[0], 5, "interval")
		default:
			return m.SetTimeout(ci.Args[0], 5, "timeout", map[string]int{"count": 3})
		}
	})
	bind("builtinTimerControl", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(timers.Load(ci.Env))
		op := must(napi.ArgAs[string](ci, 1))
		timer := &timers.Timeout{Object: napi.ToObject(ci.Args[0])}
		switch op {
		case "unref":
			if err := timer.Unref(); err != nil {
				return nil, err
			}
		case "ref":
			if err := timer.Ref(); err != nil {
				return nil, err
			}
		case "refresh":
			if err := timer.Refresh(); err != nil {
				return nil, err
			}
		case "clear":
			return nil, m.ClearTimeout(timer)
		case "clearInterval":
			return nil, m.ClearInterval(timer)
		case "clearImmediate":
			return nil, m.ClearImmediate(timer)
		}
		return napi.ValueOf(ci.Env, must(timer.HasRef()))
	})
	bind("builtinTimerPromise", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(timers.LoadPromises(ci.Env))
		return m.SetTimeout(1, "timer promise")
	})
	bind("builtinImmediatePromise", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(timers.LoadPromises(ci.Env))
		return m.SetImmediate("immediate promise")
	})
	bind("builtinGoTimer", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(timers.Load(ci.Env))
		p := must(napi.CreatePromise(ci.Env))
		_, err := m.SetTimeout(func(call *napi.CallbackInfo) (napi.ValueType, error) {
			value, err := napi.ValueOf(call.Env, "Go timer callback")
			if err != nil {
				return nil, err
			}
			return nil, p.Resolve(value)
		}, 1)
		if err != nil {
			return nil, err
		}
		return p, nil
	})
	bind("builtinStream", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(stream.Load(ci.Env))
		return m.PassThrough(map[string]any{"highWaterMark": 1})
	})
	bind("builtinStreamWrite", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		ok, err := s.Write([]byte("abc"))
		if err != nil {
			return nil, err
		}
		return napi.ValueOf(ci.Env, ok)
	})
	bind("builtinStreamEnd", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		return nil, s.End([]byte("end"))
	})
	bind("builtinStreamRead", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		return s.Read()
	})
	bind("builtinStreamPipe", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		return s.Pipe(ci.Args[1])
	})
	bind("builtinStreamListen", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		if len(ci.Args) > 2 {
			return s.Once("data", ci.Args[1])
		}
		return s.On("data", ci.Args[1])
	})
	bind("builtinReadable", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(stream.Load(ci.Env))
		return m.ReadableFrom([]string{"one", "two"})
	})
	bind("builtinStreamDestroy", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		s := &stream.Stream{Object: napi.ToObject(ci.Args[0])}
		return nil, s.Destroy()
	})
	bind("builtinDNS", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		m := must(dns.Load(ci.Env))
		return m.Lookup("localhost", map[string]any{"family": 4})
	})
	bind("builtinEvents", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		emitter := must(nodeevents.CreateEvent(ci.Env))
		listener := must(emitter.OnFunc("value", ci.Args[0]))
		_, err := emitter.EmitValues("value", 23, map[string]bool{"fromGo": true})
		if err != nil {
			return nil, err
		}
		if err := emitter.RemoveListener("value", listener); err != nil {
			return nil, err
		}
		if _, err := emitter.OnceFunc("once", ci.Args[0]); err != nil {
			return nil, err
		}
		if _, err := emitter.EmitValues("once", 24, map[string]bool{"fromGo": true}); err != nil {
			return nil, err
		}
		if _, err := emitter.EmitValues("once", 25); err != nil {
			return nil, err
		}
		return emitter, nil
	})
	bind("builtinInvalid", func(ci *napi.CallbackInfo) (napi.ValueType, error) {
		mode := must(napi.ArgAs[string](ci, 0))
		switch mode {
		case "nilModule":
			return nodejs.Wrap((*napi.Object)(nil), "nil")
		case "nilPromise":
			return nodejs.ToPromise((*napi.Object)(nil))
		case "nilBuffer":
			_, err := nodejs.CopyBytes((*napi.Buffer)(nil))
			return nil, err
		case "notPromise":
			return nodejs.ToPromise(must(napi.ValueOf(ci.Env, 23)))
		case "notBuffer":
			_, err := nodejs.CopyBytes(must(napi.ValueOf(ci.Env, "x")))
			return nil, err
		case "callback":
			_, err := nodejs.CallbackOf(ci.Env, (func())(nil))
			return nil, err
		default:
			return nil, fmt.Errorf("unknown validation case")
		}
	})
}
