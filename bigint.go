package napi

import "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"

type Bigint struct{ value }

func CreateBigint[T int64 | uint64](env EnvType, valueOf T) (*Bigint, error) {
	var value napi.Value
	var err error
	switch v := any(valueOf).(type) {
	case int64:
		if value, err = napi.MustValueErr(napi.CreateBigIntInt64(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	case uint64:
		if value, err = napi.MustValueErr(napi.CreateBigIntUint64(env.NapiValue(), v)); err != nil {
			return nil, err
		}
	}

	return &Bigint{value: &Value{env: env, valueOf: value}}, nil
}

func (big *Bigint) GetInt64() (int64, error) {
	return napi.MustValueErr2(napi.GetValueBigIntInt64(big.NapiEnv(), big.NapiValue()))
}
func (big *Bigint) GetUint64() (uint64, error) {
	return napi.MustValueErr2(napi.GetValueBigIntUint64(big.NapiEnv(), big.NapiValue()))
}
