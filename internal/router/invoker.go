package router

import (
	"errors"
	"reflect"

	"github.com/zmhuanf/feng/internal/core"
)

func CheckHandler(fn any, contextType reflect.Type) error {
	ft := reflect.TypeOf(fn)
	if ft == nil || ft.Kind() != reflect.Func {
		return errors.New("handler must be func")
	}
	if ft.NumIn() != 1 && ft.NumIn() != 2 {
		return errors.New("handler must have 1 or 2 args")
	}
	if ft.In(0) != contextType {
		return errors.New("first arg must be " + contextType.String())
	}
	if ft.NumIn() == 2 && !supportedPayloadType(ft.In(1)) {
		return errors.New("second arg must be []byte, string, bool, number, struct or slice")
	}
	if ft.NumOut() > 2 {
		return errors.New("handler must return 0, 1 or 2 values")
	}
	if ft.NumOut() == 1 && !ft.Out(0).Implements(errorType()) {
		return errors.New("single return value must be error")
	}
	if ft.NumOut() == 2 && !ft.Out(1).Implements(errorType()) {
		return errors.New("second return value must be error")
	}
	return nil
}

func Call(fn any, ctx any, data string, codec core.Codec) (string, error) {
	fv := reflect.ValueOf(fn)
	ft := fv.Type()

	params := []reflect.Value{reflect.ValueOf(ctx)}
	if ft.NumIn() > 1 {
		arg, err := decodeArg(ft.In(1), data, codec)
		if err != nil {
			return "", err
		}
		params = append(params, arg)
	}

	rets := fv.Call(params)
	switch len(rets) {
	case 0:
		return "", nil
	case 1:
		if rets[0].IsNil() {
			return "", nil
		}
		return "", rets[0].Interface().(error)
	case 2:
		if !rets[1].IsNil() {
			return "", rets[1].Interface().(error)
		}
		bytes, err := codec.Marshal(rets[0].Interface())
		if err != nil {
			return "", err
		}
		return string(bytes), nil
	default:
		return "", errors.New("unsupported return values")
	}
}

func decodeArg(argType reflect.Type, data string, codec core.Codec) (reflect.Value, error) {
	switch argType.Kind() {
	case reflect.String:
		return reflect.ValueOf(data), nil
	case reflect.Slice:
		if argType.Elem().Kind() == reflect.Uint8 {
			return reflect.ValueOf([]byte(data)), nil
		}
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.Struct, reflect.Map:
	}

	argPtr := reflect.New(argType)
	if err := codec.Unmarshal([]byte(data), argPtr.Interface()); err != nil {
		return reflect.Value{}, err
	}
	return argPtr.Elem(), nil
}

func supportedPayloadType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.String, reflect.Slice, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Float32, reflect.Float64, reflect.Struct, reflect.Map:
		return true
	default:
		return false
	}
}

func errorType() reflect.Type {
	return reflect.TypeFor[error]()
}
