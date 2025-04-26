package js

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

var (
	reflectText = reflect.TypeFor[encoding.TextMarshaler]()
	reflectJSON = reflect.TypeFor[json.Marshaler]()
	reflectTime = reflect.TypeFor[time.Time]()
)

func Reflection(env napi.EnvType, value any) (napiValue napi.ValueType, err error) {
	return reflection(env, reflect.ValueOf(value))
}

func reflection(env napi.EnvType, ptr reflect.Value) (napiValue napi.ValueType, err error) {
	defer func(err *error) {
		if err2 := recover(); err2 != nil {
			switch v := err2.(type) {
			case error:
				*err = v
			default:
				*err = fmt.Errorf("panic recover: %s", err2)
			}
		}
	}(&err)

	if ptr.IsNil() || (ptr.IsValid() && ptr.IsZero()) {
		return env.Null()
	}

	ptrType := ptr.Type()
	if ptrType.Implements(reflectText) || ptrType.Implements(reflectJSON) || ptrType.Implements(reflectTime) {
		switch v := ptr.Interface().(type) {
		case nil:
			return env.Null()
		case time.Time:
			return napi.CreateDate(env, v)
		case encoding.TextMarshaler:
			data, err := v.MarshalText()
			if err != nil {
				return nil, err
			}
			return napi.CreateString(env, string(data))
		case json.Marshaler:
			var pointData any
			data, err := v.MarshalJSON()
			if err != nil {
				return nil, err
			} else if err = json.Unmarshal(data, &pointData); err != nil {
				return nil, err
			}
			return ValueOf(env, pointData)
		}
	}

	switch ptrType.Kind() {
	case reflect.Pointer:
		return reflection(env, ptr.Elem())
	case reflect.String:
		return napi.CreateString(env, ptr.String())
	case reflect.Bool:
		return napi.CreateBoolean(env, ptr.Bool())
	case reflect.Int, reflect.Uint, reflect.Int32, reflect.Uint32, reflect.Float32, reflect.Float64:
		return napi.CreateNumberAny(env, ptr.Interface())
	case reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16:
		return napi.CreateNumberAny(env, ptr.Int())
	case reflect.Int64, reflect.Uint64:
		return napi.CreateBigint(env, ptr.Int())
	case reflect.Func:
		return funcOf(env, ptr)
	case reflect.Slice, reflect.Array:
		arr, err := napi.CreateArrayWithLength(env, ptr.Len())
		if err != nil {
			return nil, err
		}
		for index := range ptr.Len() {
			value, err := reflection(env, ptr.Index(index))
			if err != nil {
				return arr, err
			} else if err = arr.Set(index, value); err != nil {
				return arr, err
			}
		}
		return arr, nil
	case reflect.Struct:
		obj, err := napi.CreateObject(env)
		if err != nil {
			return nil, err
		}

		const propertiesTagName = "napi"
		for keyIndex := range ptrType.NumField() {
			field, fieldType := ptr.Field(keyIndex), ptrType.Field(keyIndex)
			if !fieldType.IsExported() || fieldType.Tag.Get(propertiesTagName) == "-" {
				continue
			}

			value, err := reflection(env, field)
			if err != nil {
				return obj, err
			}

			keyNamed := field.Type().Name()
			if strings.Count(fieldType.Tag.Get(propertiesTagName), ",") > 0 {
				fields := strings.SplitN(fieldType.Tag.Get(propertiesTagName), ",", 2)
				keyNamed = fields[0]
				switch fields[1] {
				case "omitempty":
					if value.IsNull() || value.IsUndefined() {
						continue
					} else if value.IsArray() {
						value, _ := napi.CreateArrayFromValue(value).Length()
						if value == 0 {
							continue
						}
					} else if value.IsString() {
						if value.ToString().String() == "" {
							continue
						}
					}
				case "omitzero":
					switch {
					case value.IsDate():
						value, _ := napi.CreateDateFromValue(value).ValueOf()
						if value == 0 {
							continue
						}
					case value.IsBigInt():
						value, _ := napi.CreateBigintFromValue(value).GetInt64()
						if value == 0 {
							continue
						}
					case value.IsNumber():
						value, _ := value.ToNumber().Int64()
						if value == 0 {
							continue
						}
					}
				}
			}
			if err = obj.Set(keyNamed, value); err != nil {
				return obj, err
			}
		}
	case reflect.Map:
	}
	return env.MustUndefined(), nil
}
