package napi

import (
	"time"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Date struct{ value }

func CreateDateFromValue(value ValueType) *Date { return &Date{value}}

func CreateDate(env EnvType, date time.Time) (*Date, error) {
	value, err := napi.MustValueErr(napi.CreateDate(env.NapiValue(), float64(date.UnixMilli())))
	if err != nil {
		return nil, err
	}
	return &Date{value: &Value{env: env, valueOf: value}}, nil
}

func (d Date) ValueOf() (float64, error) {
	return napi.MustValueErr(napi.GetDateValue(d.NapiEnv(), d.NapiValue()))
}

func (d Date) Time() (time.Time, error) {
	timeFloat, err := d.ValueOf()
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(int64(timeFloat)), nil
}
