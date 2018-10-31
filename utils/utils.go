package utils

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"runtime"
	"strings"
	"unicode"
)

func Round(f float64, n int) float64 {
	pow10 := math.Pow10(n)
	return math.Trunc((f+0.5/pow10)*pow10) / pow10
}

func GetBit(word []byte, bit uint) byte {
	return (word[bit/8]) >> (bit % 8) & 0x01
}

func MinInt(x int, y int) int {
	if x < y {
		return x
	}
	return y
}

func ConvertNumber(v interface{}) interface{} {
	switch v := v.(type) {
	case float64:
		return v
	case int64:
		return v
	case string:
		return v
	case bool:
		return v
	case int:
		return int64(v)
	case uint:
		return uint64(v)
	case uint64:
		return uint64(v)
	case []byte:
		return string(v)
	case int32:
		return int64(v)
	case int16:
		return int64(v)
	case int8:
		return int64(v)
	case uint32:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint8:
		return uint64(v)
	case float32:
		return float64(v)
	default:
		return v
	}
}

func SetField(obj interface{}, name string, value interface{}) error {
	structFieldValue := reflect.ValueOf(obj).Elem().FieldByName(name)

	if !structFieldValue.IsValid() {
		return fmt.Errorf("no such field: %s in obj", name)
	}

	if !structFieldValue.CanSet() {
		return fmt.Errorf("cannot set %s field value", name)
	}

	structFieldType := structFieldValue.Type()
	val := reflect.ValueOf(value)
	if structFieldType != val.Type() {
		return errors.New("provided value type didn't match obj field type")
	}

	structFieldValue.Set(val)
	return nil
}

func UcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToUpper(v)) + str[i+1:]
	}
	return ""
}

func LcFirst(str string) string {
	for i, v := range str {
		return string(unicode.ToLower(v)) + str[i+1:]
	}
	return ""
}

func GetLineNo() string {
	_, f, l, ok := runtime.Caller(1)
	if ok {
		fL := strings.Split(f, "/")
		f = fL[len(fL)-1]
		return fmt.Sprintf("%s:%d", f, l)
	}
	return ""
}
