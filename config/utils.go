package config

import (
	"errors"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type PropertyType string

var (
	valueTypeRegexp = regexp.MustCompile("#\\{\\w*\\}")
)

const (
	Int     PropertyType = "int"
	Bool    PropertyType = "bool"
	Float32 PropertyType = "float32"
	Float64 PropertyType = "float64"
	String  PropertyType = "string"
)

func castString(value string) (interface{}, error) {
	v, t := getValueAndType(value)
	switch t {
	case String:
		return v, nil
	case Int:
		return strconv.Atoi(v)
	case Float32:
		return strconv.ParseFloat(v, 32)
	case Float64:
		return strconv.ParseFloat(v, 64)
	case Bool:
		return strconv.ParseBool(v)
	default:
		return nil, errors.New("unknown primitive type")
	}
}

func getValueAndType(value string) (string, PropertyType) {
	t := String
	v := valueTypeRegexp.ReplaceAllStringFunc(value, func(s string) string {
		t = PropertyType(s[2 : len(s)-1])
		return ""
	})
	return v, t
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
