package napi

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const propertiesTagName = "napi"

// ValueOf converts a Go value to its corresponding N-API value representation.
// It takes an environment handle (env) and a Go value (value) of any type,
// and returns the resulting N-API value (napiValue) or an error if the conversion fails.
func ValueOf(env EnvType, value any) (napiValue ValueType, err error) {
	return valueOf(env, reflect.ValueOf(value))
}

// ValueFrom converts a N-API value (napiValue) to a Go value and stores the result in v.
// The v parameter must be a pointer to the target Go variable where the converted value will be stored.
// Returns an error if v is not a pointer or if the conversion fails.
func ValueFrom(napiValue ValueType, v any) error {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Pointer {
		return fmt.Errorf("require point to convert napi value to go value")
	}
	return valueFrom(napiValue, ptr.Elem())
}

// Convert go types to valid NAPI, if not conpatible return Undefined.
func valueOf(env EnvType, ptr reflect.Value) (napiValue ValueType, err error) {
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

	ptrType := ptr.Type()
	if ptrType.ConvertibleTo(reflect.TypeFor[ValueType]()) {
		if ptr.IsValid() {
			return ptr.Interface().(ValueType), nil
		}
		return nil, nil
	} else if !ptr.IsValid() {
		return env.Undefined()
	} else if !ptr.IsZero() && ptr.CanInterface() { // Marshalers
		switch v := ptr.Interface().(type) {
		case time.Time:
			return CreateDate(env, v)
		case encoding.TextMarshaler:
			data, err := v.MarshalText()
			if err != nil {
				return nil, err
			}
			return CreateString(env, string(data))
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
		return valueOf(env, ptr.Elem())
	case reflect.String:
		return CreateString(env, ptr.String())
	case reflect.Bool:
		return CreateBoolean(env, ptr.Bool())
	case reflect.Int, reflect.Uint, reflect.Int32, reflect.Uint32, reflect.Float32, reflect.Int8, reflect.Uint8, reflect.Int16, reflect.Uint16:
		return CreateNumber(env, ptr.Int())
	case reflect.Float64:
		return CreateNumber(env, ptr.Float())
	case reflect.Int64, reflect.Uint64:
		return CreateBigint(env, ptr.Int())
	case reflect.Func:
		return funcOf(env, ptr)
	case reflect.Slice, reflect.Array:
		arr, err := CreateArray(env, ptr.Len())
		if err != nil {
			return nil, err
		}
		for index := range ptr.Len() {
			value, err := valueOf(env, ptr.Index(index))
			if err != nil {
				return arr, err
			} else if err = arr.Set(index, value); err != nil {
				return arr, err
			}
		}
		return arr, nil
	case reflect.Struct:
		obj, err := CreateObject(env)
		if err != nil {
			return nil, err
		}

		for keyIndex := range ptrType.NumField() {
			field, fieldType := ptr.Field(keyIndex), ptrType.Field(keyIndex)
			if !fieldType.IsExported() || fieldType.Tag.Get(propertiesTagName) == "-" {
				continue
			}

			value, err := valueOf(env, field)
			if err != nil {
				return obj, err
			}

			typeof, err := value.Type()
			if err != nil {
				return nil, err
			}

			var keyNamed string
			switch v := strings.TrimSpace(fieldType.Tag.Get(propertiesTagName)); v {
			case "":
				keyNamed = fieldType.Name
			default:
				keyNamed = v
				if strings.Count(v, ",") > 0 {
					fields := strings.SplitN(v, ",", 2)
					keyNamed = fields[0]
					switch fields[1] {
					case "omitempty":
						switch typeof {
						case TypeUndefined, TypeNull, TypeUnkown:
							continue
						case TypeString:
							str, err := ToString(value).Utf8Value()
							if err != nil {
								return nil, err
							} else if str == "" {
								continue
							}
						}
					case "omitzero":
						switch typeof {
						case TypeUndefined, TypeNull, TypeUnkown:
							continue
						case TypeDate:
							value, err := ToDate(value).Time()
							if err != nil {
								return nil, err
							} else if value.Unix() == 0 {
								continue
							}
						case TypeBigInt:
							value, err := ToBigint(value).Int64()
							if err != nil {
								return nil, err
							} else if value == 0 {
								continue
							}
						case TypeNumber:
							value, err := ToNumber(value).Int()
							if err != nil {
								return nil, err
							} else if value == 0 {
								continue
							}
						case TypeArray:
							value, err := ToArray(value).Length()
							if err != nil {
								return nil, err
							} else if value == 0 {
								continue
							}
						}
					}
				}
			}
			if err = obj.Set(keyNamed, value); err != nil {
				return obj, err
			}
		}

		return obj, nil
	case reflect.Map:
		obj, err := CreateObject(env)
		if err != nil {
			return nil, err
		}
		for ptrKey, ptrValue := range ptr.Seq2() {
			key, err := valueOf(env, ptrKey)
			if err != nil {
				return nil, err
			}
			value, err := valueOf(env, ptrValue)
			if err != nil {
				return nil, err
			} else if err = obj.SetWithValue(key, value); err != nil {
				return nil, err
			}
		}
		return obj, nil
	case reflect.Interface:
		if ptr.IsValid() {
			if ptr.IsNil() {
				return env.Null()
			} else if ptr.CanInterface() {
				return valueOf(env, reflect.ValueOf(ptr.Interface()))
			}
		}
	}
	return env.Undefined()
}

// Convert javascript value to go typed value
func valueFrom(jsValue ValueType, ptr reflect.Value) error {
	typeOf, err := jsValue.Type()
	if err != nil {
		return err
	}

	ptrType := ptr.Type()
	if ptrType.ConvertibleTo(reflect.TypeFor[ValueType]()) {
		switch typeOf {
		case TypeUndefined:
			und, err := jsValue.Env().Undefined()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(und))
		case TypeNull:
			null, err := jsValue.Env().Null()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(null))
		case TypeBoolean:
			ptr.Set(reflect.ValueOf(ToBoolean(jsValue)))
		case TypeNumber:
			ptr.Set(reflect.ValueOf(ToNumber(jsValue)))
		case TypeBigInt:
			ptr.Set(reflect.ValueOf(ToBigint(jsValue)))
		case TypeString:
			ptr.Set(reflect.ValueOf(ToString(jsValue)))
		case TypeSymbol:
			// ptr.Set(reflect.ValueOf(ToSymbol(jsValue)))
		case TypeObject:
			ptr.Set(reflect.ValueOf(ToObject(jsValue)))
		case TypeFunction:
			ptr.Set(reflect.ValueOf(ToFunction(jsValue)))
		case TypeExternal:
			ptr.Set(reflect.ValueOf(ToExternal(jsValue)))
		case TypeTypedArray:
			ptr.Set(reflect.ValueOf(ToTypedArray(jsValue)))
		case TypePromise:
			// ptr.Set(reflect.ValueOf(ToPromise(jsValue)))
		case TypeBuffer:
			ptr.Set(reflect.ValueOf(ToBuffer(jsValue)))
		case TypeDate:
			ptr.Set(reflect.ValueOf(ToDate(jsValue)))
		case TypeArray:
			ptr.Set(reflect.ValueOf(ToArray(jsValue)))
		case TypeArrayBuffer:
			ptr.Set(reflect.ValueOf(ToArrayBuffer(jsValue)))
		case TypeDataView:
			ptr.Set(reflect.ValueOf(ToDataView(jsValue)))
		case TypeError:
			ptr.Set(reflect.ValueOf(ToError(jsValue)))
		}
		return nil
	}

	switch ptrType.Kind() {
	case reflect.Pointer:
		return valueFrom(jsValue, ptr.Elem())
	case reflect.Interface:
		if !ptr.CanSet() || ptrType != reflect.TypeFor[any]() {
			break
		}

		switch typeOf {
		case TypeNull, TypeUndefined, TypeUnkown:
			ptr.Set(reflect.Zero(ptrType))
			return nil
		case TypeBoolean:
			valueOf, err := ToBoolean(jsValue).Value()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(valueOf))
		case TypeNumber:
			numberValue, err := ToNumber(jsValue).Float()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(numberValue))
		case TypeBigInt:
			numberValue, err := ToBigint(jsValue).Int64()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(numberValue))
		case TypeString:
			str, err := ToString(jsValue).Utf8Value()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(str))
		case TypeDate:
			timeDate, err := ToDate(jsValue).Time()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(timeDate))
		case TypeArray:
			napiArray := ToArray(jsValue)
			size, err := napiArray.Length()
			if err != nil {
				return err
			}
			value := reflect.MakeSlice(reflect.SliceOf(ptrType), size, size)
			for index := range size {
				napiValue, err := napiArray.Get(index)
				if err != nil {
					return err
				} else if err = valueFrom(napiValue, value.Index(index)); err != nil {
					return err
				}
			}
			ptr.Set(value)
		case TypeBuffer:
			buff, err := ToBuffer(jsValue).Data()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(buff))
		case TypeObject:
			obj := ToObject(jsValue)
			goMap := reflect.MakeMap(reflect.MapOf(reflect.TypeFor[string](), reflect.TypeFor[any]()))
			for keyName, value := range obj.Seq() {
				valueOf := reflect.New(reflect.TypeFor[any]())
				if err := valueFrom(value, valueOf); err != nil {
					return err
				}
				goMap.SetMapIndex(reflect.ValueOf(keyName), valueOf)
			}
			ptr.Set(goMap)
		case TypeFunction:
			ptr.Set(reflect.ValueOf(ToFunction(jsValue)))
		}
		return nil
	case reflect.String:
		if typeOf != TypeString {
			break
		}
		valueOf, err := ToString(jsValue).Utf8Value()
		if err != nil {
			return err
		}
		ptr.Set(reflect.ValueOf(valueOf))
		return nil
	case reflect.Bool:
		if typeOf != TypeBoolean {
			break
		}
		b, err := ToBoolean(jsValue).Value()
		if err != nil {
			return err
		}
		ptr.Set(reflect.ValueOf(b))
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch typeOf {
		case TypeNumber:
			b, err := ToNumber(jsValue).Int()
			if err != nil {
				return err
			}
			ptr.SetInt(b)
			return nil
		case TypeBigInt:
			b, err := ToBigint(jsValue).Int64()
			if err != nil {
				return err
			}
			ptr.SetInt(b)
			return nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch typeOf {
		case TypeNumber:
			b, err := ToNumber(jsValue).Int()
			if err != nil {
				return err
			}
			ptr.SetUint(uint64(b))
			return nil
		case TypeBigInt:
			b, err := ToBigint(jsValue).Int64()
			if err != nil {
				return err
			}
			ptr.SetUint(uint64(b))
			return nil
		}
	case reflect.Float32, reflect.Float64:
		if typeOf != TypeNumber {
			break
		}
		f, err := ToNumber(jsValue).Float()
		if err != nil {
			return err
		}
		ptr.SetFloat(f)
		return nil
	case reflect.Func, reflect.Chan:
		return nil
	case reflect.Slice:
		if typeOf != TypeArray {
			break
		}
		jsArr := ToArray(jsValue)
		size, err := jsArr.Length()
		if err != nil {
			return err
		}
		ptr.Set(reflect.MakeSlice(ptrType, size, size))
		for index := range size {
			jsValue, err := jsArr.Get(index)
			if err != nil {
				return err
			} else if err = valueFrom(jsValue, ptr.Index(index)); err != nil {
				return err
			}
		}
	case reflect.Map:
		// Check if key is string, bool, int*, uint*, float*, else return error
		switch ptrType.Key().Kind() {
		case reflect.String:
		case reflect.Bool:
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		case reflect.Float32, reflect.Float64:
		default:
			return fmt.Errorf("cannot set Object ket to %s", ptr.Kind())
		}

		goMap := reflect.MakeMap(ptrType)
		obj := ToObject(jsValue)
		for keyName, value := range obj.Seq() {
			keySetValue := reflect.New(ptrType.Key()).Elem()
			switch ptrType.Key().Kind() {
			case reflect.String:
				keySetValue.SetString(keyName)
			case reflect.Bool:
				boolV, err := strconv.ParseBool(keyName)
				if err != nil {
					return err
				}
				keySetValue.SetBool(boolV)
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				intV, err := strconv.ParseInt(keyName, 10, 64)
				if err != nil {
					return err
				}
				keySetValue.SetInt(intV)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				intV, err := strconv.ParseUint(keyName, 10, 64)
				if err != nil {
					return err
				}
				keySetValue.SetUint(intV)
			case reflect.Float32, reflect.Float64:
				floatV, err := strconv.ParseFloat(keyName, 64)
				if err != nil {
					return err
				}
				keySetValue.SetFloat(floatV)
			}

			valueOf := reflect.New(ptrType.Elem()).Elem()
			if err := valueFrom(value, valueOf); err != nil {
				return err
			}
			goMap.SetMapIndex(keySetValue, valueOf)
		}
		ptr.Set(goMap)
		return nil
	case reflect.Struct:
		switch typeOf {
		case TypeString:
			str, err := ToString(jsValue).Utf8Value()
			if err != nil {
				return err
			}
			switch {
			case ptrType.ConvertibleTo(reflect.TypeFor[encoding.TextUnmarshaler]()):
				return ptr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(str))
			case ptrType.ConvertibleTo(reflect.TypeFor[json.Unmarshaler]()):
				var raw json.RawMessage
				if err = json.Unmarshal([]byte(str), &raw); err == nil {
					return ptr.Interface().(json.Unmarshaler).UnmarshalJSON(raw)
				}
			}
		case TypeObject:
			obj := ToObject(jsValue)
			ptr.Set(reflect.New(ptrType).Elem())
			for keyIndex := range ptrType.NumField() {
				fieldType := ptrType.Field(keyIndex)
				if !fieldType.IsExported() || fieldType.Tag.Get(propertiesTagName) == "-" {
					continue
				}

				var keyName string
				switch v := strings.TrimSpace(fieldType.Tag.Get(propertiesTagName)); v {
				case "":
					keyName = fieldType.Name
				default:
					keyName = v
					if strings.Count(fieldType.Tag.Get(propertiesTagName), ",") > 0 {
						fields := strings.SplitN(fieldType.Tag.Get(propertiesTagName), ",", 2)
						keyName = fields[0]
					}
				}
				if ok, _ := obj.Has(keyName); !ok {
					continue
				}

				value, _ := obj.Get(keyName)
				valueOf := reflect.New(fieldType.Type)
				if err := valueFrom(value, valueOf); err != nil {
					return err
				}
				ptr.Field(keyIndex).Set(valueOf.Elem())
			}
			return nil
		}
	}
	return fmt.Errorf("cannot set %s, to %s", typeOf, ptr.Kind())
}
