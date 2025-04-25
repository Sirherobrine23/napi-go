package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Error struct{ *Value }

func CreateError(env *Env, msg string) (*Error, error) {
	napiMsg, err := CreateString(env, msg)
	if err != nil {
		return nil, err
	}

	napiValue, err := napi.MustValueErr(napi.CreateError(env.NapiValue(), nil, napiMsg.NapiValue()))
	if err != nil {
		return nil, err
	}

	return &Error{
		Value: &Value{
			env:     env,
			valueOf: napiValue,
			typeof:  napi.MustValue(napi.Typeof(env.NapiValue(), napiValue)),
		},
	}, nil
}

func (er *Error) ThrowAsJavaScriptException() error {
	return napi.SingleMustValueErr(napi.Throw(er.NapiEnv(), er.NapiValue()))
}
