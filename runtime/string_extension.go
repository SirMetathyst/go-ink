package runtime

import (
	"fmt"
	"reflect"
	"strings"
)

func getTypeName(v interface{}) string {
	if t := reflect.TypeOf(v); t.Kind() == reflect.Ptr {
		return "*" + t.Elem().Name()
	} else {
		return t.Name()
	}
}

func JoinObjectsString[T any](separator string, objects []T) string {

	sb := strings.Builder{}
	isFirst := true

	for _, o := range objects {
		if !isFirst {
			sb.WriteString(separator)
		}

		if str, ok := any(o).(fmt.Stringer); ok {
			sb.WriteString(str.String())
		} else {
			sb.WriteString(getTypeName(o))
		}

		isFirst = false
	}

	return sb.String()
}

func IsStringEmpty(s string) bool {

	if len(strings.TrimSpace(s)) == 0 {
		return true
	}
	return false
}
