package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Boolean struct{ value }

func CreateBoolean(env EnvType, value bool) (*Boolean, error) {
	v, err := napi.MustValueErr(napi.GetBoolean(env.NapiValue(), value))
	if err != nil {
		return nil, err
	}

	return &Boolean{value: FromValueNapi(env, v)}, nil
}

func (bo *Boolean) ValueOf() (bool, error) {
	return napi.MustValueErr(napi.GetValueBool(bo.NapiEnv(), bo.NapiValue()))
}
