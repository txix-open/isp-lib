package http

import (
	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"time"
)

var (
	json = jsoniter.ConfigDefault
)

type JsonRestClient struct {
	c *fasthttp.Client
}

func (jrc *JsonRestClient) Invoke(method, uri string, headers map[string]string, requestBody, responsePtr interface{}) error {
	responseBody, err := jrc.do(method, uri, headers, requestBody)
	if err != nil {
		return err
	}

	if responsePtr != nil {
		if err := json.Unmarshal(responseBody, responsePtr); err != nil {
			return err
		}
	}
	return nil
}

func (jrc *JsonRestClient) InvokeWithoutHeaders(method, uri string, requestBody, responsePtr interface{}) error {
	return jrc.Invoke(method, uri, nil, requestBody, responsePtr)
}

func (jrc *JsonRestClient) Post(uri string, requestBody, responsePtr interface{}) error {
	return jrc.InvokeWithoutHeaders(POST, uri, requestBody, responsePtr)
}

func (jrc *JsonRestClient) Get(uri string, responsePtr interface{}) error {
	return jrc.InvokeWithoutHeaders(GET, uri, nil, responsePtr)
}

func (jrc *JsonRestClient) InvokeWithDynamicResponse(method, uri string, headers map[string]string, requestBody interface{}) (interface{}, error) {
	responseBody, err := jrc.do(method, uri, headers, requestBody)
	if err != nil {
		return nil, err
	}

	if len(responseBody) == 0 {
		return nil, nil
	}

	if responseBody[0] == '{' {
		var res map[string]interface{}
		if err := json.Unmarshal(responseBody, &res); err != nil {
			return nil, err
		}
		return res, nil
	} else if responseBody[0] == '[' {
		var res []interface{}
		if err := json.Unmarshal(responseBody, &res); err != nil {
			return nil, err
		}
		return res, nil
	} else {
		return map[string]string{"response": string(responseBody)}, nil
	}
}

func (jrc *JsonRestClient) do(method, uri string, headers map[string]string, requestBody interface{}) ([]byte, error) {
	body, err := prepareRequestBody(requestBody)
	if err != nil {
		return nil, err
	}

	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	prepareRequest(req, method, uri, headers, body)

	if err := jrc.c.DoTimeout(req, res, 15*time.Second); err != nil {
		return nil, err
	}

	if err := checkResponseStatusCode(res); err != nil {
		return nil, err
	}

	return res.Body(), nil
}

func NewJsonRestClient() RestClient {
	return &JsonRestClient{c: &fasthttp.Client{}}
}

func prepareRequestBody(requestBody interface{}) ([]byte, error) {
	var body []byte = nil
	if requestBody != nil {
		if bytes, err := json.Marshal(requestBody); err != nil {
			return nil, err
		} else {
			body = bytes
		}
	}
	return body, nil
}

func prepareRequest(req *fasthttp.Request, method, uri string, headers map[string]string, body []byte) {
	req.SetRequestURI(uri)
	req.Header.SetMethod(method)
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}
	if method != GET && body != nil {
		req.SetBody(body)
	}
}

func checkResponseStatusCode(res *fasthttp.Response) error {
	code := res.StatusCode()
	if code != fasthttp.StatusOK {
		return ErrorResponse{
			StatusCode: code,
			Status:     fasthttp.StatusMessage(code),
			Body:       string(res.Body()),
		}
	}
	return nil
}
