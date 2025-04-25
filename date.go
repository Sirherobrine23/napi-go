package napi

import (
	"time"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Date struct{ *Value }

func CreateDate(env *Env, date time.Time) (*Date, error) {
	value, err := napi.MustValueErr(napi.CreateDate(env.NapiValue(), float64(date.UnixMilli())))
	if err != nil {
		return nil, err
	}

	return &Date{
		Value: &Value{
			env:     env,
			valueOf: value,
			typeof:  napi.MustValue(napi.Typeof(env.NapiValue(), value)),
		},
	}, nil
}

func (d Date) ValueOf() (float64, error) {
	return napi.MustValueErr(napi.GetDateValue(d.env.env, d.Value.valueOf))
}

func (d Date) Time() (time.Time, error) {
	timeFloat, err := d.ValueOf()
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(int64(timeFloat)), nil
}
