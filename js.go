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
func ValueFrom(napiValue ValueType, v any) (err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			if cause, ok := recovered.(error); ok {
				err = fmt.Errorf("decode JavaScript value: %w", cause)
			} else {
				err = fmt.Errorf("decode JavaScript value: %v", recovered)
			}
		}
	}()
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Pointer || ptr.IsNil() {
		return fmt.Errorf("ValueFrom requires a non-nil pointer")
	}
	if nilValue(napiValue) || napiValue.NapiValue() == nil {
		return fmt.Errorf("cannot decode a nil JavaScript value")
	}
	return valueFrom(napiValue, ptr.Elem())
}

// Convert go types to valid NAPI, if not conpatible return Undefined.
func valueOf(env EnvType, ptr reflect.Value) (napiValue ValueType, err error) {
	return valueOfDepth(env, ptr, 0)
}

var maxConversionDepth = 256

func valueOfDepth(env EnvType, ptr reflect.Value, depth int) (napiValue ValueType, err error) {
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

	if depth > maxConversionDepth {
		return nil, fmt.Errorf("Go value exceeds conversion depth %d (possibly cyclic)", maxConversionDepth)
	}
	if !ptr.IsValid() {
		return env.Undefined()
	}
	ptrType := ptr.Type()
	if ptrType.ConvertibleTo(reflect.TypeFor[ValueType]()) {
		value := ptr.Interface().(ValueType)
		if nilValue(value) {
			return env.Undefined()
		}
		return value, validateValue(env, value)
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
			return valueOfDepth(env, reflect.ValueOf(pointData), depth+1)
		}
	}

	switch ptrType.Kind() {
	case reflect.Pointer:
		return valueOfDepth(env, ptr.Elem(), depth+1)
	case reflect.String:
		return CreateString(env, ptr.String())
	case reflect.Bool:
		return CreateBoolean(env, ptr.Bool())
	case reflect.Int, reflect.Int32, reflect.Int8, reflect.Int16:
		return CreateNumber(env, ptr.Int())
	case reflect.Uint, reflect.Uint32, reflect.Uint8, reflect.Uint16:
		return CreateNumber(env, ptr.Uint())
	case reflect.Float32, reflect.Float64:
		return CreateNumber(env, ptr.Float())
	case reflect.Int64:
		return CreateBigint(env, ptr.Int())
	case reflect.Uint64:
		return CreateBigint(env, ptr.Uint())
	case reflect.Func:
		if ptr.IsNil() {
			return env.Undefined()
		}
		return funcOf(env, ptr)
	case reflect.Slice, reflect.Array:
		arr, err := CreateArray(env, ptr.Len())
		if err != nil {
			return nil, err
		}
		for index := range ptr.Len() {
			value, err := valueOfDepth(env, ptr.Index(index), depth+1)
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

			value, err := valueOfDepth(env, field, depth+1)
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
			if keyNamed == "" {
				keyNamed = fieldType.Name
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
			key, err := valueOfDepth(env, ptrKey, depth+1)
			if err != nil {
				return nil, err
			}
			value, err := valueOfDepth(env, ptrValue, depth+1)
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
				return valueOfDepth(env, ptr.Elem(), depth+1)
			}
		}
	}
	return env.Undefined()
}

// Convert javascript value to go typed value
func valueFrom(jsValue ValueType, ptr reflect.Value) error {
	return valueFromDepth(jsValue, ptr, 0)
}

func valueFromDepth(jsValue ValueType, ptr reflect.Value, depth int) error {
	if depth > maxConversionDepth {
		return fmt.Errorf("JavaScript value exceeds conversion depth %d (possibly cyclic)", maxConversionDepth)
	}
	if !ptr.IsValid() || !ptr.CanSet() {
		return fmt.Errorf("decode target is not settable")
	}
	typeOf, err := jsValue.Type()
	if err != nil {
		return err
	}

	ptrType := ptr.Type()
	if ptrType.Implements(reflect.TypeFor[ValueType]()) {
		wrapped := wrapDecodedValue(jsValue, typeOf)
		if ptrType == reflect.TypeFor[*Class]() && typeOf == TypeFunction {
			wrapped = ToClass(jsValue)
		}
		converted := reflect.ValueOf(wrapped)
		if converted.Type().AssignableTo(ptrType) {
			ptr.Set(converted)
			return nil
		}
		if converted.Kind() == reflect.Pointer && converted.Elem().Type().AssignableTo(ptrType) {
			ptr.Set(converted.Elem())
			return nil
		}
		return fmt.Errorf("cannot decode %s as %s", typeOf, ptrType)
	}

	switch ptrType.Kind() {
	case reflect.Pointer:
		if typeOf == TypeNull || typeOf == TypeUndefined {
			ptr.SetZero()
			return nil
		}
		if ptr.IsNil() {
			ptr.Set(reflect.New(ptrType.Elem()))
		}
		return valueFromDepth(jsValue, ptr.Elem(), depth+1)
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
				} else if err = valueFromDepth(napiValue, value.Index(index), depth+1); err != nil {
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
				valueOf := reflect.New(reflect.TypeFor[any]()).Elem()
				if err := valueFromDepth(value, valueOf, depth+1); err != nil {
					return err
				}
				goMap.SetMapIndex(reflect.ValueOf(keyName), valueOf)
			}
			ptr.Set(goMap)
		default:
			ptr.Set(reflect.ValueOf(wrapDecodedValue(jsValue, typeOf)))
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
		ptr.SetString(valueOf)
		return nil
	case reflect.Bool:
		if typeOf != TypeBoolean {
			break
		}
		b, err := ToBoolean(jsValue).Value()
		if err != nil {
			return err
		}
		ptr.SetBool(b)
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
		return fmt.Errorf("cannot decode %s as %s", typeOf, ptrType)
	case reflect.Slice:
		if typeOf == TypeNull || typeOf == TypeUndefined {
			ptr.SetZero()
			return nil
		}
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
			} else if err = valueFromDepth(jsValue, ptr.Index(index), depth+1); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		if typeOf == TypeNull || typeOf == TypeUndefined {
			ptr.SetZero()
			return nil
		}
		if typeOf != TypeObject {
			break
		}
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
			if err := valueFromDepth(value, valueOf, depth+1); err != nil {
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
				if keyName == "" {
					keyName = fieldType.Name
				}
				ok, err := obj.Has(keyName)
				if err != nil {
					return err
				}
				if !ok {
					continue
				}
				value, err := obj.Get(keyName)
				if err != nil {
					return err
				}
				valueOf := reflect.New(fieldType.Type).Elem()
				if err := valueFromDepth(value, valueOf, depth+1); err != nil {
					return err
				}
				ptr.Field(keyIndex).Set(valueOf)
			}
			return nil
		}
	}
	return fmt.Errorf("cannot set %s, to %s", typeOf, ptr.Kind())
}
