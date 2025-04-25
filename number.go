package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Number struct{ value }

func CreateNumber[T int32 | uint32 | int64 | float64](env EnvType, n T) (*Number, error) {
	var value napi.Value
	var err error
	switch v := any(n).(type) {
	case int32:
		if value, err = napi.MustValueErr(napi.CreateInt32(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	case uint32:
		if value, err = napi.MustValueErr(napi.CreateUint32(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	case int64:
		if value, err = napi.MustValueErr(napi.CreateInt64(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	case float64:
		if value, err = napi.MustValueErr(napi.CreateDouble(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	}

	return &Number{value: FromValueNapi(env, value)}, err
}

func (num *Number) Int64() (int64, error) {
	return napi.MustValueErr(napi.GetValueInt64(num.NapiEnv(), num.NapiValue()))
}

func (num *Number) Int32() (int32, error) {
	return napi.MustValueErr(napi.GetValueInt32(num.NapiEnv(), num.NapiValue()))
}

func (num *Number) Float64() (float64, error) {
	return napi.MustValueErr(napi.GetValueDouble(num.NapiEnv(), num.NapiValue()))
}
