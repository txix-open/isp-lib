package http

import "fmt"

type content struct {
	uriPart     string
	filePath    string
	contentType string
	bytes       []byte
	renderParam map[string]string
}

func (c content) String() string {
	return fmt.Sprintf("%s -> %s", c.uriPart, c.filePath)
}

func Serve(uriPart, filePath string, renderParams ...string) *content {
	return &content{uriPart, filePath, "", []byte{}, toMap(renderParams)}
}

func ServeV2(uriPart, filePath, contentType string, renderParams ...string) *content {
	return &content{uriPart, filePath, contentType, []byte{}, toMap(renderParams)}
}

func toMap(renderParams []string) map[string]string {
	m := make(map[string]string)
	if len(renderParams) == 0 {
		return m
	}
	if len(renderParams)%2 != 0 {
		renderParams = renderParams[:len(renderParams)-1]
	}
	if len(renderParams) == 0 {
		return m
	}
	for i := 0; i < len(renderParams)-1; i++ {
		m[renderParams[i]] = renderParams[i+1]
	}
	return m
}
