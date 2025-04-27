package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Boolean struct{ value }

// Convert [ValueType] to [*Boolean]
func ToBoolean(o ValueType) *Boolean { return &Boolean{o} }

func CreateBoolean(env EnvType, value bool) (*Boolean, error) {
	v, err := mustValueErr(napi.GetBoolean(env.NapiValue(), value))
	if err != nil {
		return nil, err
	}
	return ToBoolean(N_APIValue(env, v)), nil
}

func (bo *Boolean) Value() (bool, error) {
	return mustValueErr(napi.GetValueBool(bo.NapiEnv(), bo.NapiValue()))
}
