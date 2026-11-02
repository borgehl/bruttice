package util

import (
	"reflect"
	"regexp"
	"strings"
)

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

// Returns snake case of v
func StructToSnakeCase(v any) string {
	vType := reflect.TypeOf(v)
	for vType.Kind() == reflect.Ptr {
		vType = vType.Elem()
	}
	snake := matchFirstCap.ReplaceAllString(vType.Name(), "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
