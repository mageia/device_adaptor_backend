package utils

import (
	"errors"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

type OffsetBitPair [][3]interface{}	//[offset, bit, "key"]

func (c OffsetBitPair) Len() int {
	return len(c)
}
func (c OffsetBitPair) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c OffsetBitPair) Less(i, j int) bool {
	return c[i][0].(int) < c[j][0].(int)
}
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

func IsExists(p string) bool {
	_, err := os.Stat(p)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsExported(name string) bool {
	ch, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(ch)
}

func extractTag(tag string) (col, rest string) {
	tags := strings.SplitN(tag, ",", 2)
	if len(tags) == 2 {
		return strings.TrimSpace(tags[0]), strings.TrimSpace(tags[1])
	}
	return strings.TrimSpace(tags[0]), ""
}
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	result := make([]rune, 0, len(s))
	upper := false
	for _, r := range s {
		if r == '_' {
			upper = true
			continue
		}
		if upper {
			result = append(result, unicode.ToUpper(r))
			upper = false
			continue
		}
		result = append(result, r)
	}
	result[0] = unicode.ToUpper(result[0])
	return string(result)
}

func FindField(rv reflect.Value, name string) (field reflect.Value, fieldName string, found bool) {
	switch rv.Kind() {
	case reflect.Struct:
		rt := rv.Type()
		for i := 0; i < rt.NumField(); i++ {
			ft := rt.Field(i)
			if !IsExported(ft.Name) {
				continue
			}
			if col, _ := extractTag(ft.Tag.Get("json")); col == name {
				return rv.Field(i), ft.Name, true
			}
			if ft.Anonymous {
				rva := rv.Field(i)
				field, name, found := FindField(rva, name)
				if found {
					return field, name, found
				}
			}
		}
		for _, name := range []string{
			strings.Title(name),
			toCamelCase(name),
			strings.ToUpper(name),
		} {
			if field := rv.FieldByName(name); field.IsValid() {
				return field, name, true
			}
		}
	case reflect.Map:
		return reflect.New(rv.Type().Elem()).Elem(), name, true
	}
	return field, "", false
}
