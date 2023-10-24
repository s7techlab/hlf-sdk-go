package transform

import (
	"bytes"
	"strings"
)

const (
	nullByte            = "\x00"
	alternativeNullByte = "\\u0000"
	keySeparator        = `_`
)

var (
	byteNullByte            = []byte(nullByte)
	byteAlternativeNullByte = []byte(`\\u0000`)
)

func ReplaceBytesU0000ToNullBytes(str []byte) []byte {
	if str == nil {
		return nil
	}

	return bytes.ReplaceAll(str, byteAlternativeNullByte, byteNullByte)
}

func ReplaceStringSeparatorToNullBytes(str string) string {
	if len(str) == 0 {
		return ""
	}
	return strings.ReplaceAll(str, keySeparator, nullByte)
}

func ReplaceStringNullBytesToSeparator(str string) string {
	if len(str) == 0 {
		return ""
	}
	return strings.ReplaceAll(str, nullByte, keySeparator)
}

func RemoveStringNullBytes(str string) string {
	if len(str) == 0 {
		return ""
	}
	str = strings.ReplaceAll(str, nullByte, ``)
	return strings.ReplaceAll(str, alternativeNullByte, ``)
}

func RemoveBytesNullBytes(str []byte) []byte {
	if str == nil {
		return nil
	}
	str = bytes.ReplaceAll(str, byteNullByte, []byte{})
	return bytes.ReplaceAll(str, byteAlternativeNullByte, []byte{})
}
