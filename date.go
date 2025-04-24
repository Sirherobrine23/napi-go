package napi

import (
	"time"

	"sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

type Date struct{ *Value }

func (d Date) ValueOf() float64 { return napi.MustValue(napi.GetDateValue(d.env.env, d.Value.valueOf)) }

func (d Date) GoTime() time.Time { return time.UnixMilli(int64(d.ValueOf())) }
