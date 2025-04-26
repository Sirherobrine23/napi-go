package js

import (
	"fmt"
	"reflect"
	"strings"

	"sirherobrine23.com.br/Sirherobrine23/napi-go"
)

func reflectFromTypedValue(env napi.EnvType, goPtr reflect.Value, napiValue napi.ValueType) error {
	switch goPtr.Kind() {
	case reflect.Ptr:
		if napiValue.IsUndefined() || napiValue.IsNull() {
			goPtr.Set(reflect.Zero(goPtr.Type()))
		} else {
			if goPtr.IsNil() {
				goPtr.Set(reflect.New(goPtr.Type().Elem()))
			}
			return reflectFromTypedValue(env, goPtr.Elem(), napiValue)
		}
	case reflect.String:
		if napiValue.IsString() {
			str, err := napiValue.ToString().ValueOf()
			if err != nil {
				return err
			}
			goPtr.SetString(str)
			return nil
		}
		return fmt.Errorf("return mismatch value from napi value, expected string return %s", napiValue.Type())
	case reflect.Bool:
		if napiValue.IsBoolean() {
			b, err := napiValue.ToBoolean().ValueOf()
			if err != nil {
				return err
			}
			goPtr.SetBool(b)
			return nil
		}
		return fmt.Errorf("return mismatch value from napi value, expected bool return %s", napiValue.Type())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if napiValue.IsNumber() {
			n, err := napiValue.ToNumber().Int64()
			if err != nil {
				return err
			}
			goPtr.SetInt(n)
			return nil
		} else if napiValue.IsBigInt() {
			n, err := napi.CreateBigintFromValue(napiValue).GetInt64()
			if err != nil {
				return err
			}
			goPtr.SetInt(n)
			return nil
		}
		return fmt.Errorf("return mismatch value from napi value, expected %s return %s", goPtr.Kind(), napiValue.Type())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if napiValue.IsNumber() {
			n, err := napiValue.ToNumber().Int64()
			if err != nil {
				return err
			}
			goPtr.SetUint(uint64(n))
			return nil
		} else if napiValue.IsBigInt() {
			n, err := napi.CreateBigintFromValue(napiValue).GetUint64()
			if err != nil {
				return err
			}
			goPtr.SetUint(uint64(n))
			return nil
		}
		return fmt.Errorf("return mismatch value from napi value, expected %s return %s", goPtr.Kind(), napiValue.Type())
	case reflect.Float32, reflect.Float64:
		if napiValue.IsNumber() {
			n, err := napiValue.ToNumber().Float64()
			if err != nil {
				return err
			}
			goPtr.SetFloat(n)
			return nil
		} else if napiValue.IsBigInt() {
			n, err := napi.CreateBigintFromValue(napiValue).GetInt64()
			if err != nil {
				return err
			}
			goPtr.SetFloat(float64(n))
			return nil
		}
		return fmt.Errorf("return mismatch value from napi value, expected %s return %s", goPtr.Kind(), napiValue.Type())
	case reflect.Slice:
		if napiValue.IsArray() {
			napiArray := napi.CreateArrayFromValue(napiValue)
			arrayInt, err := napiArray.Length()
			if err != nil {
				return err
			}
			if goPtr.IsNil() {
				goPtr.Set(reflect.MakeSlice(goPtr.Type(), arrayInt, arrayInt))
			}
			for i := range arrayInt {
				napiValue, err := napiArray.Get(i)
				if err != nil {
					return err
				} else if err := reflectFromTypedValue(env, goPtr.Index(i), napiValue); err != nil {
					return err
				}
			}
			return nil
		}
	case reflect.Array:
		if napiValue.IsArray() {
			napiArray := napi.CreateArrayFromValue(napiValue)
			goPtr.Set(reflect.MakeSlice(goPtr.Type(), goPtr.Len(), goPtr.Len()))
			for index := range goPtr.Len() {
				napiValue, err := napiArray.Get(index)
				if err != nil {
					break
				} else if err = reflectFromTypedValue(env, goPtr.Index(index), napiValue); err != nil {
					return err
				}
			}
			return nil
		}
		return fmt.Errorf("return mismatch value from napi value, expected %s return %s", goPtr.Kind(), napiValue.Type())
	case reflect.Map:
		if napiValue.IsObject() {
			napiObject := napiValue.ToObject()
			if goPtr.IsNil() {
				goPtr.Set(reflect.MakeMap(goPtr.Type()))
			}
			for keyName, napiValue := range napiObject.Seq() {
				if err := reflectFromTypedValue(env, goPtr.MapIndex(reflect.ValueOf(keyName)), napiValue); err != nil {
					return err
				}
			}
			return nil
		}
		return fmt.Errorf("return mismatch value from napi value, expected %s return %s", goPtr.Kind(), napiValue.Type())
	case reflect.Struct:
		if napiValue.IsObject() {
			napiObject := napiValue.ToObject()
			if goPtr.IsNil() {
				goPtr.Set(reflect.New(goPtr.Type()).Elem())
			}
			ptrType := goPtr.Type()
			const propertiesTagName = "napi"
			for keyIndex := range ptrType.NumField() {
				field, fieldType := goPtr.Field(keyIndex), ptrType.Field(keyIndex)
				if !fieldType.IsExported() || fieldType.Tag.Get(propertiesTagName) == "-" {
					continue
				}
				keyNamed, omitEmpty, omitZero := field.Type().Name(), false, false
				if strings.Count(fieldType.Tag.Get(propertiesTagName), ",") > 0 {
					fields := strings.SplitN(fieldType.Tag.Get(propertiesTagName), ",", 2)
					keyNamed = fields[0]
					switch fields[1] {
					case "omitempty":
						omitEmpty = true
					case "omitzero":
						omitZero = true
					}
				}

				napiValue, err := napiObject.Get(keyNamed)
				if err != nil {
					return err
				}
				if napiValue.IsUndefined() || napiValue.IsNull() {
					if omitEmpty || omitZero {
						continue
					}
				}
				if err := reflectFromTypedValue(env, field, napiValue); err != nil {
					return err
				}
			}
			return nil
		}
		return fmt.Errorf("return mismatch value from napi value, expected %s return %s", goPtr.Kind(), napiValue.Type())
	case reflect.Interface:
		// if is any/interface{} return map[string]interface{}
		if goPtr.IsValid() && goPtr.IsZero() {
			switch napiValue.Type() {
			case napi.TypeUndefined, napi.TypeNull:
				goPtr.Set(reflect.Zero(goPtr.Type()))
			case napi.TypeBoolean:
				b, err := napiValue.ToBoolean().ValueOf()
				if err != nil {
					return err
				}
				goPtr.Set(reflect.ValueOf(b))
			case napi.TypeNumber:
				n, err := napiValue.ToNumber().Float64()
				if err != nil {
					return err
				}
				goPtr.Set(reflect.ValueOf(n))
			case napi.TypeBigInt:
				n, err := napiValue.ToNumber().Int64()
				if err != nil {
					return err
				}
				goPtr.Set(reflect.ValueOf(n))
			case napi.TypeString:
				str, err := napiValue.ToString().ValueOf()
				if err != nil {
					return err
				}
				goPtr.SetString(str)
			case napi.TypeObject:
				// napiObject := napiValue.ToObject()
			case napi.TypeBuffer:
				// napi.Create()
			case napi.TypeArray:
				arr := napi.CreateArrayFromValue(napiValue)
				size, err := arr.Length()
				if err != nil {
					return err
				}
				goPtr.Set(reflect.MakeSlice(reflect.TypeFor[[]any](), size, size))
				for i := range size {
					napiValue, err := arr.Get(i)
					if err != nil {
						return err
					} else if err = reflectFromTypedValue(env, goPtr.Index(i), napiValue); err != nil {
						return err
					}
				}
			}
			return nil
		}
	}
	return nil
}
