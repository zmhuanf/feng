package client

import (
	"errors"
	"reflect"
)

func call(f any, ctx IContext, data []byte) ([]byte, error) {
	// 验证参数合法性
	fv := reflect.ValueOf(f)
	ft := reflect.TypeOf(f)
	if ft.Kind() != reflect.Func {
		return nil, errors.New("f must be func")
	}
	if ft.NumIn() != 2 {
		return nil, errors.New("func must have 2 args")
	}
	if ft.In(0) != reflect.TypeOf((*IContext)(nil)).Elem() {
		return nil, errors.New("first arg must be IContext")
	}

	// 根据参数类型解析参数
	argType := reflect.TypeOf(f).In(1)
	var argValue reflect.Value
	switch argType.Kind() {
	case reflect.Struct:
		// 接收结构体参数
		argPtr := reflect.New(argType)
		err := ctx.GetClient().GetConfig().Codec.Unmarshal(data, argPtr.Interface())
		if err != nil {
			return nil, err
		}
		argValue = argPtr.Elem()
	case reflect.String:
		// 接受字符串参数
		argValue = reflect.ValueOf(string(data))
	case reflect.Slice:
		// 接受字节数组参数
		if argType.Elem().Kind() != reflect.Uint8 {
			return nil, errors.New("slice arg must be []byte")
		}
		argValue = reflect.ValueOf(data)
	default:
		return nil, errors.New("unsupported arg type")
	}
	// 调用
	rets := fv.Call([]reflect.Value{reflect.ValueOf(ctx), argValue})
	// 有返回值，处理返回值
	if len(rets) > 0 {
		// 处理err类型
		if rets[0].Type().Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			if !rets[0].IsNil() {
				return nil, rets[0].Interface().(error)
			}
			return nil, nil
		}
		result := rets[0].Interface()
		resultBytes, err := ctx.GetClient().GetConfig().Codec.Marshal(result)
		if err != nil {
			return nil, err
		}
		return resultBytes, nil
	}
	return nil, nil
}
