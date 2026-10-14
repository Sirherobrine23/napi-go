// Package nodejs provides scoped wrappers around the actual Node.js builtins.
// All operations require the owning Node thread and a live handle scope. Load
// modules inside each callback; do not cache their scoped handles across calls
// or share them between Workers. Use TSFN/EventEmitterThread for Go producers.
package nodejs

import (
	"fmt"
	"reflect"
	"strings"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	internal "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/nodejs/require"
)

// Module exposes both typed helpers and the complete module through Object.
// A returned Promise or other object remains a JavaScript-owned scoped handle.
type Module struct {
	*napi.Object
	ID string
}

// Load synchronously loads a builtin, with or without its node: prefix.
// Unknown/unavailable modules return an error; there is no process-global cache.
func Load(env napi.EnvType, id string) (*Module, error) {
	if !strings.HasPrefix(id, "node:") {
		id = "node:" + id
	}
	value, err := require.Builtin(env, id)
	if err != nil {
		return nil, err
	}
	return Wrap(value, id)
}

// Wrap binds an existing module or namespace in its current callback scope.
func Wrap(value napi.ValueType, id string) (*Module, error) {
	if err := checkValue(value); err != nil {
		return nil, fmt.Errorf("module %q: %w", id, err)
	}
	object, err := napi.ObjectOf(value.Env(), value)
	if err != nil {
		return nil, err
	}
	return &Module{Object: object, ID: id}, nil
}

// Call invokes a module method with Go arguments, preserving the module as this.
func (module *Module) Call(name string, args ...any) (napi.ValueType, error) {
	value, err := module.CallMethod(name, args...)
	if err != nil {
		return nil, fmt.Errorf("%s.%s: %w", module.ID, name, err)
	}
	return value, nil
}

func (module *Module) CallInto(name string, target any, args ...any) error {
	value, err := module.Call(name, args...)
	if err != nil {
		return err
	}
	if err := napi.ValueFrom(value, target); err != nil {
		return fmt.Errorf("%s.%s result: %w", module.ID, name, err)
	}
	return nil
}

func CallAs[T any](module *Module, name string, args ...any) (T, error) {
	var result T
	err := module.CallInto(name, &result, args...)
	return result, err
}

// New constructs an exported JavaScript class with Go arguments.
func (module *Module) New(name string, args ...any) (*napi.Object, error) {
	class, err := napi.GetAs[*napi.Class](module.Object, name)
	if err != nil {
		return nil, err
	}
	return class.NewValues(args...)
}

// CallbackOf accepts a Go callback/function or an existing JS function.
func CallbackOf(env napi.EnvType, callback any) (*napi.Function, error) {
	var value napi.ValueType
	var err error
	if existing, ok := callback.(napi.ValueType); ok {
		value, err = napi.ValueOf(env, existing)
	} else {
		value, err = napi.GoFuncOf(env, callback)
	}
	if err != nil {
		return nil, err
	}
	kind, err := value.Type()
	if err != nil {
		return nil, err
	}
	if kind != napi.TypeFunction {
		return nil, fmt.Errorf("callback must be a function, got %s", kind)
	}
	return napi.ToFunction(value), nil
}

// Binary copies Go []byte to a Node Buffer instead of ValueOf's JS array. Other
// inputs keep normal ValueOf semantics (strings or existing binary JS handles).
func Binary(env napi.EnvType, input any) (napi.ValueType, error) {
	if data, ok := input.([]byte); ok {
		return napi.CopyBuffer(env, data)
	}
	return napi.ValueOf(env, input)
}

// CopyBytes copies a Node Buffer into Go-owned memory before the handle expires.
// The resulting slice can be sent to background Go code. This method itself
// must run on the Node thread. It does not retain a view of JavaScript memory.
func CopyBytes(value napi.ValueType) ([]byte, error) {
	if err := checkValue(value); err != nil {
		return nil, err
	}
	ok, status := internal.IsBuffer(value.NapiEnv(), value.NapiValue())
	if err := status.ToError(); err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("expected a Node Buffer")
	}
	data, err := napi.ToBuffer(value).Data()
	if err != nil {
		return nil, err
	}
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

func checkValue(value napi.ValueType) error {
	if value == nil {
		return fmt.Errorf("JavaScript value is nil")
	}
	ptr := reflect.ValueOf(value)
	if ptr.Kind() == reflect.Pointer && ptr.IsNil() {
		return fmt.Errorf("JavaScript value is nil")
	}
	if value.NapiEnv() == nil || value.NapiValue() == nil {
		return fmt.Errorf("JavaScript value has no environment or handle")
	}
	return nil
}

// Strings adapts variadic strings to Go-valued call helpers.
func Strings(values []string) []any {
	args := make([]any, len(values))
	for i, value := range values {
		args[i] = value
	}
	return args
}
