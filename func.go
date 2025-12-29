package feng

import (
	"errors"
	"reflect"
)

func call(f any, c any, data string) (string, error) {
	fv := reflect.ValueOf(f)
	ft := reflect.TypeOf(f)

	var codec ICodec
	switch ctx := c.(type) {
	case IServerContext:
		codec = ctx.GetServer().GetConfig().Codec
	case IClientContext:
		codec = ctx.GetClient().GetConfig().Codec
	default:
		return "", errors.New("invalid context type")
	}

	// 参数
	params := []reflect.Value{reflect.ValueOf(c)}
	if ft.NumIn() > 1 {
		argType := reflect.TypeOf(f).In(1)
		var argValue reflect.Value
		switch argType.Kind() {
		case reflect.Struct:
			// 接收结构体参数
			argPtr := reflect.New(argType)
			err := codec.Unmarshal([]byte(data), argPtr.Interface())
			if err != nil {
				return "", err
			}
			argValue = argPtr.Elem()
		case reflect.String:
			// 接受字符串参数
			argValue = reflect.ValueOf(string(data))
		case reflect.Slice:
			// 接受字节数组参数
			if argType.Elem().Kind() != reflect.Uint8 {
				return "", errors.New("slice arg must be []byte")
			}
			argValue = reflect.ValueOf(data)
		default:
			return "", errors.New("unsupported arg type")
		}
		params = append(params, argValue)
	}
	// 调用
	rets := fv.Call(params)
	// 有返回值，处理返回值
	switch len(rets) {
	case 0:
		// 没有返回值，直接返回
		return "", nil
	case 1:
		// 只有一个返回值，是error
		if !rets[0].IsNil() {
			return "", rets[0].Interface().(error)
		}
		return "", nil
	case 2:
		// 有两个返回值，第一个是结果，第二个是error
		if !rets[1].IsNil() {
			return "", rets[1].Interface().(error)
		}
		result := rets[0].Interface()
		resultBytes, err := codec.Marshal(result)
		if err != nil {
			return "", err
		}
		return string(resultBytes), nil
	default:
		return "", errors.New("unsupported return values")
	}
}
