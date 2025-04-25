package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Boolean struct{ *Value }

func CreateBoolean(env *Env, value bool) (*Boolean, error) {
	v, err := napi.MustValueErr(napi.GetBoolean(env.NapiValue(), value))
	if err != nil {
		return nil, err
	}

	return &Boolean{
		Value: &Value{
			env:     env,
			typeof:  napi.ValueTypeBoolean,
			valueOf: v,
		},
	}, nil
}

func (bo *Boolean) ValueOf() (bool, error) {
	return napi.MustValueErr(napi.GetValueBool(bo.NapiEnv(), bo.NapiValue()))
}
