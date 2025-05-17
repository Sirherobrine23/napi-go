package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Boolean struct{ value }

// Convert [ValueType] to [*Boolean]
func ToBoolean(o ValueType) *Boolean { return &Boolean{o} }

// CreateBoolean creates a new Boolean value in the given N-API environment.
// It takes an EnvType representing the N-API environment and a Go bool value.
// Returns a pointer to a Boolean object representing the value in the N-API environment,
// or an error if the creation fails.
func CreateBoolean(env EnvType, value bool) (*Boolean, error) {
	v, err := mustValueErr(napi.GetBoolean(env.NapiValue(), value))
	if err != nil {
		return nil, err
	}
	return ToBoolean(N_APIValue(env, v)), nil
}

// Value retrieves the boolean value represented by the Boolean object.
// It returns the Go bool value and an error if the underlying N-API call fails.
func (bo *Boolean) Value() (bool, error) {
	return mustValueErr(napi.GetValueBool(bo.NapiEnv(), bo.NapiValue()))
}
