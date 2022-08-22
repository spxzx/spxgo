package bytesconv

import "unsafe"

// StringToBytes 不会发生内存拷贝，性能相对较好
func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(
		&struct {
			string
			Cap int
		}{s, len(s)},
	))
}
