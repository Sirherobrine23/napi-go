package main

import (
	"time"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	"sirherobrine23.com.br/Sirherobrine23/napi-go/entry"
)

func init() {
	entry.Register(func(e napi.EnvType, o *napi.Object) {
		value, _ := napi.CreateString(e, "test")
		o.Set("test", value)

		fn, _ := napi.CreateFunction(e, "testFunc", func(env napi.EnvType, this napi.ValueType, args []napi.ValueType) (napi.ValueType, error) {
			if this.IsObject() {
				for keyName := range this.ToObject().Seq() {
					println(keyName)
				}
			}

			return napi.MustCreateString(env, "test"), nil
		})
		o.Set("testFunc", fn)

		fnDate, _ := napi.CreateFunction(e, "testFunc", func(env napi.EnvType, this napi.ValueType, args []napi.ValueType) (napi.ValueType, error) {
			current := time.Now().UTC()
			println(current.Format(time.RFC3339))
			return napi.CreateDate(env, current)
		})
		o.Set("testFuncDate", fnDate)
	})
}

func main() {}
