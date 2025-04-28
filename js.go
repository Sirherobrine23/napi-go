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

// Convert go types to valid NAPI, if not conpatible return Undefined.
func ValueOf(env EnvType, value any) (napiValue ValueType, err error) {
	return valueOf(env, reflect.ValueOf(value))
}

// Convert NAPI value to Go values
func ValueFrom(napiValue ValueType, v any) error {
	ptr := reflect.ValueOf(v)
	if ptr.Kind() != reflect.Pointer {
		return fmt.Errorf("require point to convert napi value to go value")
	}
	return valueFrom(napiValue, ptr.Elem())
}

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

			keyNamed := fieldType.Name
			if strings.Count(fieldType.Tag.Get(propertiesTagName), ",") > 0 {
				fields := strings.SplitN(fieldType.Tag.Get(propertiesTagName), ",", 2)
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
			// ptr.Set(reflect.ValueOf(ToString(jsValue)))
		case TypeObject:
			ptr.Set(reflect.ValueOf(ToObject(jsValue)))
		case TypeFunction:
			ptr.Set(reflect.ValueOf(ToFunction(jsValue)))
		case TypeExternal:
			// ptr.Set(reflect.ValueOf(ToFunction(jsValue)))
		case TypeTypedArray:
			// ptr.Set(reflect.ValueOf(ToFunction(jsValue)))
		case TypePromise:
			// ptr.Set(reflect.ValueOf(ToFunction(jsValue)))
		case TypeDataView:
			// ptr.Set(reflect.ValueOf(ToFunction(jsValue)))
		case TypeBuffer:
			ptr.Set(reflect.ValueOf(ToBuffer(jsValue)))
		case TypeDate:
			ptr.Set(reflect.ValueOf(ToDate(jsValue)))
		case TypeArray:
			ptr.Set(reflect.ValueOf(ToArray(jsValue)))
		case TypeArrayBuffer:
			// ptr.Set(reflect.ValueOf(ToArray(jsValue)))
		case TypeError:
			ptr.Set(reflect.ValueOf(ToError(jsValue)))
		}
		return nil
	}

	switch ptrType.Kind() {
	case reflect.Pointer:
		return valueFrom(jsValue, ptr.Elem())
	case reflect.Interface:
		// Check if is any and can set
		if ptr.CanSet() && ptrType == reflect.TypeFor[any]() {
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
		}
		return fmt.Errorf("cannot set value, returned %s", typeOf)
	}

	switch typeOf {
	case TypeNull, TypeUndefined, TypeUnkown:
		switch ptrType.Kind() {
		case reflect.Interface, reflect.Pointer:
			ptr.Set(reflect.Zero(ptrType))
			return nil
		default:
			return fmt.Errorf("cannot set value, returned %s", typeOf)
		}
	case TypeBoolean:
		switch ptr.Kind() {
		case reflect.Bool:
			valueOf, err := ToBoolean(jsValue).Value()
			if err != nil {
				return err
			}
			ptr.SetBool(valueOf)
		default:
			return fmt.Errorf("cannot set boolean value to %s", ptr.Kind())
		}
	case TypeNumber:
		switch ptrType.Kind() {
		case reflect.Float32, reflect.Float64:
			floatValue, err := ToNumber(jsValue).Float()
			if err != nil {
				return err
			}
			ptr.SetFloat(floatValue)
			return nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			numberValue, err := ToNumber(jsValue).Int()
			if err != nil {
				return err
			}
			ptr.SetInt(numberValue)
			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			numberValue, err := ToNumber(jsValue).Int()
			if err != nil {
				return err
			}
			ptr.SetUint(uint64(numberValue))
			return nil
		default:
			return fmt.Errorf("cannot set number value to %s", ptr.Kind())
		}
	case TypeBigInt:
		switch ptrType.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			numberValue, err := ToNumber(jsValue).Int()
			if err != nil {
				return err
			}
			ptr.SetInt(numberValue)
			return nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			numberValue, err := ToNumber(jsValue).Int()
			if err != nil {
				return err
			}
			ptr.SetUint(uint64(numberValue))
			return nil
		default:
			return fmt.Errorf("cannot set number value to %s", ptr.Kind())
		}
	case TypeString:
		str, err := ToString(jsValue).Utf8Value()
		if err != nil {
			return err
		}
		switch {
		case ptr.CanConvert(reflect.TypeFor[encoding.TextUnmarshaler]()):
			return ptr.Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(str))
		case ptr.Kind() == reflect.String:
			ptr.Set(reflect.ValueOf(str))
			return nil
		}
		return fmt.Errorf("cannot set string to %s", ptr.Kind())
	case TypeDate:
		if ptrType == reflect.TypeFor[time.Time]() {
			timeDate, err := ToDate(jsValue).Time()
			if err != nil {
				return err
			}
			ptr.Set(reflect.ValueOf(timeDate))
			return nil
		}
		return fmt.Errorf("cannot set Date to %s", ptr.Kind())
	case TypeArray:
		napiArray := ToArray(jsValue)
		size, err := napiArray.Length()
		if err != nil {
			return err
		}

		switch ptr.Kind() {
		case reflect.Slice:
			value := reflect.MakeSlice(ptrType, size, size)
			for index := range size {
				napiValue, err := napiArray.Get(index)
				if err != nil {
					return err
				} else if err = valueFrom(napiValue, value.Index(index)); err != nil {
					return err
				}
			}
			ptr.Set(value)
			return nil
		case reflect.Array:
			value := reflect.New(ptrType)
			for index := range min(size, value.Len()) {
				napiValue, err := napiArray.Get(index)
				if err != nil {
					return err
				} else if err = valueFrom(napiValue, value.Index(index)); err != nil {
					return err
				}
			}
			ptr.Set(value)
			return nil
		default:
			return fmt.Errorf("cannot set Array to %s", ptr.Kind())
		}
	case TypeBuffer:
		if ptrType == reflect.TypeFor[[]byte]() {
			buff, err := ToBuffer(jsValue).Data()
			if err != nil {
				return err
			}
			ptr.SetBytes(buff)
			return nil
		}
		return fmt.Errorf("cannot set Buffer to %s", ptr.Kind())
	case TypeObject:
		obj := ToObject(jsValue)
		switch ptr.Kind() {
		case reflect.Struct:
			ptr.Set(reflect.New(ptrType).Elem())
			for keyIndex := range ptrType.NumField() {
				field, fieldType := ptr.Field(keyIndex), ptrType.Field(keyIndex)
				if !fieldType.IsExported() || fieldType.Tag.Get(propertiesTagName) == "-" {
					continue
				}

				keyName, omitEmpty, omitZero := fieldType.Name, false, false
				if strings.Count(fieldType.Tag.Get(propertiesTagName), ",") > 0 {
					fields := strings.SplitN(fieldType.Tag.Get(propertiesTagName), ",", 2)
					keyName = fields[0]
					switch fields[1] {
					case "omitempty":
						omitEmpty = true
					case "omitzero":
						omitZero = true
					}
				} else {
					omitEmpty, omitZero = true, true
				}

				if ok, err := obj.Has(keyName); err != nil {
					return err
				} else if !ok && !(omitEmpty || omitZero) {
					return fmt.Errorf("cannot set %s to %s", keyName, ptr.Kind())
				}

				value, err := obj.Get(keyName)
				if err != nil {
					return err
				}

				valueTypeof, _ := value.Type()
				if omitEmpty || omitZero {
					switch valueTypeof {
					case TypeUndefined, TypeNull, TypeUnkown:
						continue
					case TypeString:
						if str, _ := ToString(value).Utf8Value(); str == "" {
							continue
						}
					case TypeDate:
						if timeDate, _ := ToDate(value).Time(); timeDate.Unix() == 0 {
							continue
						}
					case TypeBigInt:
						if numberValue, _ := ToBigint(value).Int64(); numberValue == 0 {
							continue
						}
					case TypeNumber:
						if numberValue, _ := ToNumber(value).Int(); numberValue == 0 {
							continue
						}
					case TypeArray:
						if size, _ := ToArray(value).Length(); size == 0 {
							continue
						}
					}
				}

				valueOf := reflect.New(fieldType.Type).Elem()
				if err := valueFrom(value, valueOf); err != nil {
					return err
				}
				field.Set(valueOf)
			}
			return nil
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
		default:
			return fmt.Errorf("cannot set Object to %s", ptr.Kind())
		}
	}

	return nil
}
