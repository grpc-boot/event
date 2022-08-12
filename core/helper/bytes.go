package helper

import "unsafe"

func Bytes2String(data []byte) string {
	return *(*string)(unsafe.Pointer(&data))
}
