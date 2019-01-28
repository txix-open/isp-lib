package config

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func castString(resultType reflect.Type, value string) (interface{}, error) {
	switch resultType.Kind() {
	case reflect.String:
		return value, nil
	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return strconv.Atoi(value)
	case reflect.Float32:
		return strconv.ParseFloat(value, 32)
	case reflect.Float64:
		return strconv.ParseFloat(value, 64)
	case reflect.Bool:
		return strconv.ParseBool(value)
	default:
		return nil, errors.New("unknown primitive type")
	}
}

func getEnvOverrides(envPrefix string) map[string]string {
	m := make(map[string]string)

	vars := os.Environ()
	for _, v := range vars {
		pairs := strings.Split(v, "=")
		if len(pairs) >= 2 && strings.HasPrefix(pairs[0], envPrefix) {
			path := pairs[0][len(envPrefix):]
			path = strings.ToLower(path)
			m[path] = pairs[1]
		}
	}

	return m
}
