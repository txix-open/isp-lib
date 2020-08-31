package docs

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"

	"github.com/swaggo/swag"
)

type SwaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

type s struct {
	doc   string
	sInfo SwaggerInfo
}

func (s *s) ReadDoc() string {
	s.sInfo.Description = strings.Replace(s.sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(s.doc)
	if err != nil {
		return s.doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, s.sInfo); err != nil {
		return s.doc
	}

	return tpl.String()
}

func InitSwagger(SwaggerInfo SwaggerInfo, doc string) {
	swag.Register(swag.Name, &s{doc: doc, sInfo: SwaggerInfo})
}
