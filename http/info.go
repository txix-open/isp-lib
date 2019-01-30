package http

import (
	"fmt"
	"sort"
	"strings"
)

type HandlersInfoSnapshot struct {
	soapHandlers   []string
	restHandlers   []string
	staticHandlers []string
}

func (s HandlersInfoSnapshot) String() string {
	builder := &strings.Builder{}
	builder.WriteRune('\n')
	write := func(header string, ss []string) {
		builder.WriteString(header)
		builder.WriteRune('\n')
		if len(ss) == 0 {
			builder.WriteString("\t---EMPTY---\n")
			return
		}
		for _, h := range ss {
			builder.WriteRune('\t')
			builder.WriteString(h)
			builder.WriteRune('\n')
		}
	}
	write("REST", s.restHandlers)
	write("SOAP", s.soapHandlers)
	write("STATIC", s.staticHandlers)
	builder.WriteRune('\n')
	return builder.String()
}

func makeSnapshot(actionMap map[string]*funcDesc, staticMap map[string]*content) HandlersInfoSnapshot {
	soap := make([]string, 0)
	rest := make([]string, 0)
	static := make([]string, 0)
	for key, fd := range actionMap {
		if fd.mType == SoapMType {
			soap = append(soap, fmt.Sprintf("POST %s soapAction:%q -> %s", fd.uri, fd.method, fd.String()))
		} else if fd.mType == RestMType {
			rest = append(rest, fmt.Sprintf("POST %s -> %s", key, fd.String()))
		}
	}
	for _, c := range staticMap {
		static = append(static, fmt.Sprintf("GET %s", c.String()))
	}
	sort.Strings(soap)
	sort.Strings(rest)
	sort.Strings(static)
	return HandlersInfoSnapshot{soap, rest, static}
}
