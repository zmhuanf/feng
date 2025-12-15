package feng

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
)

func sign(message, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return hex.EncodeToString(h.Sum(nil))
}

func verify(message, secret, signature string) bool {
	expectedSignature := sign(message, secret)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

func generateRandomKey(length int) string {
	key := make([]byte, length)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(key)
}

func checkFuncTypeServer(fn any) error {
	// 检查函数签名
	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Func {
		return errors.New("f must be func")
	}
	if ft.NumIn() != 2 {
		return errors.New("func must have 2 args")
	}
	if ft.In(0) != reflect.TypeFor[IServerContext]() {
		return errors.New("first arg must be IServerContext")
	}
	secondArg := ft.In(1)
	switch secondArg.Kind() {
	case reflect.Slice:
		if secondArg.Elem().Kind() != reflect.Uint8 {
			return errors.New("second arg must be []byte, string or struct")
		}
	case reflect.String:
	case reflect.Struct:
	default:
		return errors.New("second arg must be []byte, string or struct")
	}
	return nil
}

func checkFuncTypeClient(fn any) error {
	// 检查函数签名
	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Func {
		return errors.New("f must be func")
	}
	if ft.NumIn() != 2 {
		return errors.New("func must have 2 args")
	}
	if ft.In(0) != reflect.TypeFor[IClientContext]() {
		return errors.New("first arg must be IClientContext")
	}
	secondArg := ft.In(1)
	switch secondArg.Kind() {
	case reflect.Slice:
		if secondArg.Elem().Kind() != reflect.Uint8 {
			return errors.New("second arg must be []byte, string or struct")
		}
	case reflect.String:
	case reflect.Struct:
	default:
		return errors.New("second arg must be []byte, string or struct")
	}
	return nil
}
