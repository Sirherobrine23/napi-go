// Package os queries the Node host. Results follow Node's platform and Worker context.
package os

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs"
)

type Module struct{ *nodejs.Module }
type CPUInfo struct {
	Model string   `napi:"model"`
	Speed int      `napi:"speed"`
	Times CPUTimes `napi:"times"`
}
type CPUTimes struct {
	User float64 `napi:"user"`
	Nice float64 `napi:"nice"`
	Sys  float64 `napi:"sys"`
	Idle float64 `napi:"idle"`
	IRQ  float64 `napi:"irq"`
}
type UserInfo struct {
	UID      int     `napi:"uid"`
	GID      int     `napi:"gid"`
	Username string  `napi:"username"`
	Homedir  string  `napi:"homedir"`
	Shell    *string `napi:"shell"`
}

func Load(env napi.EnvType) (*Module, error) {
	m, err := nodejs.Load(env, "node:os")
	if err != nil {
		return nil, err
	}
	return &Module{m}, nil
}
func (m *Module) CPUs() ([]CPUInfo, error) { return nodejs.CallAs[[]CPUInfo](m.Module, "cpus") }
func (m *Module) UserInfo(options ...any) (UserInfo, error) {
	return nodejs.CallAs[UserInfo](m.Module, "userInfo", options...)
}
func (m *Module) NetworkInterfaces() (*napi.Object, error) {
	value, err := m.Call("networkInterfaces")
	if err != nil {
		return nil, err
	}
	return napi.ObjectOf(m.Env(), value)
}
func (m *Module) Arch() (string, error)       { return nodejs.CallAs[string](m.Module, "arch") }
func (m *Module) Platform() (string, error)   { return nodejs.CallAs[string](m.Module, "platform") }
func (m *Module) Hostname() (string, error)   { return nodejs.CallAs[string](m.Module, "hostname") }
func (m *Module) OSType() (string, error)     { return nodejs.CallAs[string](m.Module, "type") }
func (m *Module) Release() (string, error)    { return nodejs.CallAs[string](m.Module, "release") }
func (m *Module) Homedir() (string, error)    { return nodejs.CallAs[string](m.Module, "homedir") }
func (m *Module) Tmpdir() (string, error)     { return nodejs.CallAs[string](m.Module, "tmpdir") }
func (m *Module) Endianness() (string, error) { return nodejs.CallAs[string](m.Module, "endianness") }
func (m *Module) Uptime() (float64, error)    { return nodejs.CallAs[float64](m.Module, "uptime") }
func (m *Module) Totalmem() (float64, error)  { return nodejs.CallAs[float64](m.Module, "totalmem") }
func (m *Module) Freemem() (float64, error)   { return nodejs.CallAs[float64](m.Module, "freemem") }
func (m *Module) Loadavg() ([]float64, error) { return nodejs.CallAs[[]float64](m.Module, "loadavg") }
func (m *Module) AvailableParallelism() (int, error) {
	return nodejs.CallAs[int](m.Module, "availableParallelism")
}
