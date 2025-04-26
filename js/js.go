// Implementation bind to Go types to N-API struct
package js

import (
	"sirherobrine23.com.br/Sirherobrine23/napi-go"
	internalNapi "sirherobrine23.com.br/Sirherobrine23/napi-go/internal/napi"
)

// Convert go types to valid NAPI, if not conpatible return Undefined.
func ValueOf(env napi.EnvType, value any) (napi.ValueType, error) {
	switch v := value.(type) {
	case napi.ValueType:
		return v, nil
	case internalNapi.Value:
		return napi.FromValueNapi(env, v), nil
	case []internalNapi.Value:
		arr, err := napi.CreateArray(env)
		if err == nil {
			for _, v := range v {
				if err = arr.Append(napi.FromValueNapi(env, v)); err != nil {
					break
				}
			}
		}
		return arr, err
	case []napi.ValueType:
		arr, err := napi.CreateArray(env)
		if err != nil {
			return nil, err
		}
		return arr, arr.Append(v...)
	case int:
		return napi.CreateNumber(env, v)
	case uint:
		return napi.CreateNumber(env, v)
	case int32:
		return napi.CreateNumber(env, v)
	case uint32:
		return napi.CreateNumber(env, v)
	case int64:
		return napi.CreateNumber(env, v)
	case uint64:
		return napi.CreateNumber(env, v)
	case float64:
		return napi.CreateNumber(env, v)
	case string:
		return napi.CreateString(env, v)
	case nil:
		return env.Null()
	case bool:
		return napi.CreateBoolean(env, v)
	case error:
		if v == nil {
			return env.Undefined()
		}
		return napi.CreateError(env, v.Error())
	case []any:
		arr, err := napi.CreateArrayWithLength(env, len(v))
		if err != nil {
			return nil, err
		}
		var value napi.ValueType
		for sliceIndex, preValue := range v {
			if value, err = ValueOf(env, preValue); err != nil {
				return arr, err
			}
			arr.Set(sliceIndex, value)
		}
		return arr, err
	case map[string]any:
		obj, err := napi.CreateObject(env)
		if err != nil {
			return nil, err
		}

		var value napi.ValueType
		for name, preValue := range v {
			if value, err = ValueOf(env, preValue); err != nil {
				return obj, err
			}
			obj.Set(name, value)
		}
		return obj, err
	case napi.Callback, internalNapi.Callback:
		return GoFuncOf(env, v)
	default:
		return Reflection(env, v)
	}
}
