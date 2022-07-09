package formula

import (
	"errors"
	"fmt"
	"math/bits"
	"reflect"
	"strconv"
)

var ErrNotSupported = errors.New("Type not supported")

type Decoder struct{}

func (fd *Decoder) Decode(target any, src map[string][]string) error {
	value := reflect.ValueOf(target)
	kind := value.Kind()

	if value.Kind() == reflect.Pointer {
		kind = value.Elem().Kind()
		value = value.Elem()
	} else {
		return fmt.Errorf("Decode not implemented for: %s", value.Kind())
	}

	switch kind {
	case reflect.Struct:
		err := decodeStruct(target, value, src)
		return err
	default:
		return fmt.Errorf("Decode not implemented for: %s", value.Kind())
	}
}

func decodeStruct(target any, value reflect.Value, src map[string][]string) error {
	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)

		if !field.CanSet() {
			continue
		}

		structField := value.Type().Field(i)
		fieldName := structField.Name

		if tag, ok := structField.Tag.Lookup("param"); ok {
			fieldName = tag
		}
		fieldValue, ok := src[fieldName]

		if !ok {
			continue
		}

		fieldType := field.Type()
		fieldKind := fieldType.Kind()

		if fieldKind == reflect.Pointer {
			fieldKind = fieldType.Elem().Kind()
		}

		switch fieldKind {
		case reflect.String,
			reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Bool:
			parsedVal, err := decodeSingle(fieldType, fieldValue[0])

			if err != nil {
				return err
			}

			field.Set(parsedVal)
		case reflect.Slice:
			slice, err := decodeSlice(fieldType, fieldValue)
			if err != nil {
				return err
			}

			if fieldType.Kind() == reflect.Pointer {
				ptr := reflect.New(fieldType.Elem())
				ptr.Elem().Set(slice)

				field.Set(ptr)
			} else {
				field.Set(slice)
			}
		default:
		}
	}

	return nil
}

func decodeSlice(fieldType reflect.Type, values []string) (reflect.Value, error) {
	var slice reflect.Value
	if fieldType.Kind() == reflect.Pointer {
		slice = reflect.MakeSlice(fieldType.Elem(), 0, len(values))
	} else {
		slice = reflect.MakeSlice(fieldType, 0, len(values))
	}

	sliceElem := sliceElemType(fieldType)

	for _, value := range values {
		decodedValue, err := decodeSingle(sliceElem, value)
		if err != nil {
			continue
		}
		slice = reflect.Append(slice, decodedValue)
	}

	return slice, nil
}

func sliceElemType(fieldType reflect.Type) reflect.Type {
	if fieldType.Kind() == reflect.Pointer {
		fieldType = fieldType.Elem()
	}

	return fieldType.Elem()
}

func decodeSingle(fieldType reflect.Type, value string) (reflect.Value, error) {
	kind := fieldType.Kind()

	// Handle non-pointer happy path
	if kind != reflect.Pointer {
		decodedValue, err := decodeSingleNonPointer(kind, value)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(decodedValue), err
	}

	realKind := fieldType.Elem().Kind()

	if realKind == reflect.Pointer {
		return reflect.Value{}, ErrNotSupported
	}

	decodedValue, err := decodeSingleNonPointer(realKind, value)
	if err != nil {
		return reflect.Value{}, err
	}

	ptr := reflect.New(fieldType.Elem())
	ptr.Elem().Set(reflect.ValueOf(decodedValue))

	return ptr, nil
}

func decodeSingleNonPointer(kind reflect.Kind, value string) (any, error) {
	switch kind {
	case reflect.String:
		return value, nil
	case reflect.Bool:
		return strconv.ParseBool(value)
	case reflect.Int:
		return strconv.Atoi(value)
	case reflect.Int8:
		value, err := strconv.ParseInt(value, 10, 8)
		return int8(value), err
	case reflect.Int16:
		value, err := strconv.ParseInt(value, 10, 16)
		return int16(value), err
	case reflect.Int32:
		value, err := strconv.ParseInt(value, 10, 32)
		return int32(value), err
	case reflect.Int64:
		return strconv.ParseInt(value, 10, 64)
	case reflect.Uint:
		value, err := strconv.ParseUint(value, 10, bits.UintSize)
		return uint(value), err
	case reflect.Uint8:
		value, err := strconv.ParseUint(value, 10, 8)
		return uint8(value), err
	case reflect.Uint16:
		value, err := strconv.ParseUint(value, 10, 16)
		return uint16(value), err
	case reflect.Uint32:
		value, err := strconv.ParseUint(value, 10, 32)
		return uint32(value), err
	case reflect.Uint64:
		return strconv.ParseUint(value, 10, 64)
	case reflect.Float32:
		value, err := strconv.ParseFloat(value, 32)

		if err != nil {
			return nil, err
		}

		return float32(value), nil
	case reflect.Float64:
		return strconv.ParseFloat(value, 64)
	case reflect.Complex64:
		value, err := strconv.ParseComplex(value, 64)

		if err != nil {
			return nil, err
		}

		return complex64(value), nil
	case reflect.Complex128:
		return strconv.ParseComplex(value, 128)
	default:
		return nil, ErrNotSupported
	}
}
