package utils

import "unsafe"

// String 字节数组转字符串
func String(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Bytes 字符串转字节数组
func Bytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
