package x

import (
	jsoniter "github.com/json-iterator/go"
	"unsafe"
)

var Json = jsoniter.ConfigFastest

func MustUnmarshal(b []byte, v interface{}) error {
	_ = Json.Unmarshal(b, v)
	return nil
}

func MustMarshal(v interface{}) []byte {
	b, _ := Json.Marshal(v)
	return b
}

func MustMarshal2String(v interface{}) string {
	b, _ := Json.Marshal(v)
	return BytesToString(b)
}

func StringToBytes(s string) (b []byte) {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
