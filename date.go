package napi

import (
	"time"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Date struct{ value }

// Convert [ValueType] to [*Date]
func ToDate(o ValueType) *Date { return &Date{o} }

func CreateDate(env EnvType, date time.Time) (*Date, error) {
	value, err := mustValueErr(napi.CreateDate(env.NapiValue(), float64(date.UnixMilli())))
	if err != nil {
		return nil, err
	}
	return &Date{value: &Value{env: env, valueOf: value}}, nil
}

// Get time from [*Date] object.
func (d Date) Time() (time.Time, error) {
	timeFloat, err := mustValueErr(napi.GetDateValue(d.NapiEnv(), d.NapiValue()))
	if err != nil {
		return time.Time{}, err
	}
	return time.UnixMilli(int64(timeFloat)), nil
}
