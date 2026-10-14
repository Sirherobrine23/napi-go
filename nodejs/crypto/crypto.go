// Package crypto wraps Node/OpenSSL hashing and random bytes without substituting
// Go algorithms. Availability and errors follow the running Node build.
package crypto

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }

// Hash also wraps Hmac, which supports the same update/digest methods.
type Hash struct{ *napi.Object }

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:crypto")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) CreateHash(algorithm string, options ...any) (*Hash, error) {
	value, err := m.Call("createHash", append([]any{algorithm}, options...)...)
	if err != nil {
		return nil, err
	}
	object, err := napi.ObjectOf(m.Env(), value)
	if err != nil {
		return nil, err
	}
	return &Hash{object}, nil
}
func (m *Module) CreateHmac(algorithm string, key any, options ...any) (*Hash, error) {
	keyValue, err := nodejs.Binary(m.Env(), key)
	if err != nil {
		return nil, err
	}
	value, err := m.Call("createHmac", append([]any{algorithm, keyValue}, options...)...)
	if err != nil {
		return nil, err
	}
	object, err := napi.ObjectOf(m.Env(), value)
	if err != nil {
		return nil, err
	}
	return &Hash{object}, nil
}
func (h *Hash) Update(data any, encoding ...string) error {
	value, err := nodejs.Binary(h.Env(), data)
	if err != nil {
		return err
	}
	_, err = h.CallMethod("update", append([]any{value}, nodejs.Strings(encoding)...)...)
	return err
}
func (h *Hash) Digest(encoding ...string) (napi.ValueType, error) {
	return h.CallMethod("digest", nodejs.Strings(encoding)...)
}
func (h *Hash) DigestBytes() ([]byte, error) {
	value, err := h.Digest()
	if err != nil {
		return nil, err
	}
	return nodejs.CopyBytes(value)
}
func (h *Hash) DigestHex() (string, error) {
	value, err := h.Digest("hex")
	if err != nil {
		return "", err
	}
	return napi.Decode[string](value)
}
func (m *Module) GetHashes() ([]string, error) { return nodejs.CallAs[[]string](m.Module, "getHashes") }
func (m *Module) RandomBytes(size int) ([]byte, error) {
	value, err := m.Call("randomBytes", size)
	if err != nil {
		return nil, err
	}
	return nodejs.CopyBytes(value)
}
func (m *Module) RandomUUID(options ...any) (string, error) {
	return nodejs.CallAs[string](m.Module, "randomUUID", options...)
}
func (m *Module) TimingSafeEqual(a, b []byte) (bool, error) {
	lhs, err := nodejs.Binary(m.Env(), a)
	if err != nil {
		return false, err
	}
	rhs, err := nodejs.Binary(m.Env(), b)
	if err != nil {
		return false, err
	}
	return nodejs.CallAs[bool](m.Module, "timingSafeEqual", lhs, rhs)
}
