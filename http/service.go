package http

import (
	"encoding/xml"
	"fmt"
	"github.com/integration-system/gowsdl/soap"
	"github.com/integration-system/isp-lib/logger"
	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	soapActionHeader = "SOAPAction"
	xmlContentType   = `application/xml; charset="utf-8"`
	jsonContentType  = `application/json; charset="utf-8"`
)

var (
	json = jsoniter.ConfigFastest
)

type RESTFault struct {
	Code   int
	Status string
}

func (rf *RESTFault) Error() string {
	return fmt.Sprintf("%d %s", rf.Code, rf.Status)
}

type HttpService struct {
	server            *fasthttp.Server
	actions           map[string]*funcDesc
	static            map[string]*content
	errorMapper       ErrorMapper
	unimplErrorMapper UnimplMethodErrorMapper
	mws               []Middleware
	pp                []func(ctx *Ctx)
}

func (ss *HttpService) Register(uri, method string, mType MType, handler interface{}) error {
	if desc, err := toDesc(handler); err != nil {
		return err
	} else {
		desc.mType = mType
		desc.method = method
		desc.uri = uri
		ss.actions[toKey(uri, method)] = desc
		return nil
	}
}

func (ss *HttpService) RegisterControllers(uri string, handlers ...interface{}) error {
	err := registerControllers(ss.Register, uri, handlers)
	if err != nil {
		ss.actions = map[string]*funcDesc{}
		return err
	} else {
		return nil
	}
}

func (ss *HttpService) RegisterStatic(list ...*content) error {
	for _, c := range list {
		f, err := os.Open(c.filePath)
		if err != nil {
			if f != nil {
				_ = f.Close()
			}
			return err
		}
		bytes, err := ioutil.ReadAll(f)
		if err != nil {
			_ = f.Close()
			return err
		}
		_ = f.Close()
		c.bytes = bytes
		ss.static[c.uriPart] = c
	}
	return nil
}

func (ss *HttpService) GetHandlersSnapshot() HandlersInfoSnapshot {
	return makeSnapshot(ss.actions, ss.static)
}

func (ss *HttpService) ListenAndServe(bindingAddress string) error {
	ln, err := net.Listen("tcp", bindingAddress)
	if err != nil {
		return err
	}
	return ss.Serve(ln)
}

func (ss *HttpService) Serve(ln net.Listener) error {
	return ss.server.Serve(ln)
}

func (ss *HttpService) Shutdown() error {
	return ss.server.Shutdown()
}

func (ss *HttpService) handleRequest(ctx *fasthttp.RequestCtx) {
	uri := string(ctx.Request.URI().RequestURI())
	if c, ok := ss.static[uri]; ok && ctx.IsGet() {
		if c.contentType != "" {
			ctx.SetContentType(c.contentType)
		} else {
			ctx.SetContentType(xmlContentType)
		}
		ctx.SetStatusCode(http.StatusOK)
		ctx.Write(c.bytes)
		return
	}

	key := getActionKey(ctx)
	fd, ok := ss.actions[key]
	c := &Ctx{RequestCtx: ctx, m: make(map[string]interface{}), action: key}
	if ok {
		if fd.mType == SoapMType {
			ss.handleSoapRequest(fd, c)
		} else if fd.mType == RestMType {
			ss.handleRestRequest(fd, c)
		}
	} else {
		ct := string(c.Request.Header.ContentType())
		if strings.HasPrefix(ct, "application/xml") || strings.HasPrefix(ct, "text/xml") {
			handled := false
			if ss.unimplErrorMapper != nil {
				content := ss.unimplErrorMapper(c, key)
				if content != nil {
					respBody := soap.SOAPEnvelope{Body: soap.SOAPBody{Content: content}}
					writeXmlBody(c, respBody)
					handled = true
				}
			}
			if !handled {
				c.SetStatusCode(http.StatusInternalServerError)
				respBody := soap.SOAPEnvelope{Body: soap.SOAPBody{
					Fault: &soap.SOAPFault{Code: "501", String: fmt.Sprintf("%s - not implemented", key)},
				}}
				writeXmlBody(c, respBody)
			}
		} else if strings.HasPrefix(ct, "application/json") {
			handled := false
			if ss.unimplErrorMapper != nil {
				content := ss.unimplErrorMapper(c, key)
				if content != nil {
					writeJsonBody(c, content)
					handled = true
				}
			}
			if !handled {
				c.SetStatusCode(http.StatusNotImplemented)
				respBody := RESTFault{Code: http.StatusNotImplemented, Status: fmt.Sprintf("%s - not implemented", key)}
				writeJsonBody(c, respBody)
			}

		} else {
			c.SetStatusCode(http.StatusNotImplemented)
		}
	}

	for _, p := range ss.pp {
		p(c)
	}
}

func (ss *HttpService) handleRestRequest(fd *funcDesc, ctx *Ctx) {
	var err error
	var result interface{}
	for _, mw := range ss.mws {
		if ctx, err = mw(ctx); err != nil {
			break
		}
	}

	if err == nil {
		result, err = handleRestRequest(fd, ctx)
	}

	if err != nil {
		ctx.err = err
		if ss.errorMapper != nil {
			result = ss.errorMapper(ctx, err)
		} else if fault, ok := err.(*RESTFault); ok {
			result = fault
			ctx.SetStatusCode(fault.Code)
		} else {
			result = &RESTFault{
				Code:   http.StatusInternalServerError,
				Status: http.StatusText(http.StatusInternalServerError),
			}
			ctx.SetStatusCode(http.StatusInternalServerError)
			logger.Warn(err)
		}
	} else {
		ctx.SetStatusCode(http.StatusOK)
	}

	if result != nil {
		writeJsonBody(ctx, result)
	}
}

func (ss *HttpService) handleSoapRequest(fd *funcDesc, ctx *Ctx) {
	var err error
	var result interface{}
	respBody := soap.SOAPEnvelope{}
	for _, mw := range ss.mws {
		if ctx, err = mw(ctx); err != nil {
			break
		}
	}

	if err == nil {
		result, err = handleSoapRequest(fd, ctx)
	}

	if err != nil {
		ctx.err = err
		if ss.errorMapper != nil {
			respBody.Body = soap.SOAPBody{Content: ss.errorMapper(ctx, err)}
		} else if fault, ok := err.(*soap.SOAPFault); ok {
			respBody.Body = soap.SOAPBody{Fault: fault}
			ctx.SetStatusCode(http.StatusInternalServerError)
		} else {
			respBody.Body = soap.SOAPBody{Fault: &soap.SOAPFault{Code: "500", String: "Internal service error"}}
			ctx.SetStatusCode(http.StatusInternalServerError)
			logger.Warn(err)
		}
	} else if result != nil {
		respBody.Body = soap.SOAPBody{Content: result}
		ctx.SetStatusCode(http.StatusOK)
	}

	writeXmlBody(ctx, respBody)
}

func NewService(opts ...Option) *HttpService {
	ss := &HttpService{
		actions: make(map[string]*funcDesc),
		static:  make(map[string]*content),
		mws:     []Middleware{},
		pp:      []func(c *Ctx){},
	}
	server := &fasthttp.Server{
		Handler: ss.handleRequest,
	}
	ss.server = server
	for _, o := range opts {
		o(ss)
	}
	return ss
}

func handleSoapRequest(fd *funcDesc, ctx *Ctx) (interface{}, error) {
	params := make([]reflect.Value, fd.inCount)
	reqBody := &soap.SOAPEnvelope{}
	if fd.bodyNum != -1 {
		val := reflect.New(fd.inType)
		reqBody.Body = soap.SOAPBody{Content: val.Interface()}
		params[fd.bodyNum] = val
	}

	if fd.ctxNum != -1 {
		params[fd.ctxNum] = reflect.ValueOf(ctx)
	}

	if fd.headersNum != -1 {
		val := reflect.New(reflect.TypeOf([]interface{}{}))
		//reqBody.Header = soap.SOAPHeader{Items: val.Interface()}
		params[fd.headersNum] = val
	}

	err := xml.Unmarshal(ctx.PostBody(), reqBody)
	if err != nil {
		logger.Warn(err)
		return nil, &soap.SOAPFault{Code: "400", String: "Invalid xml request body"}
	}
	ctx.mappedRequestBody = reflect.ValueOf(reqBody.Body.Content).Elem().Interface()
	//todo add validation
	return callF(fd, params)
}

func handleRestRequest(fd *funcDesc, ctx *Ctx) (interface{}, error) {
	params := make([]reflect.Value, fd.inCount)
	if fd.bodyNum != -1 {
		val := reflect.New(fd.inType)
		err := json.Unmarshal(ctx.PostBody(), val.Interface())
		if err != nil {
			logger.Warn(err)
			return nil, &RESTFault{Code: http.StatusBadRequest, Status: "Invalid json request body"}
		}
		//todo add validation
		ctx.mappedRequestBody = val.Elem().Interface()
		params[fd.bodyNum] = val
	}

	if fd.ctxNum != -1 {
		params[fd.ctxNum] = reflect.ValueOf(ctx)
	}

	if fd.headersNum != -1 {
		val := reflect.New(reflect.TypeOf([]interface{}{}))
		//reqBody.Header = soap.SOAPHeader{Items: val.Interface()}
		params[fd.headersNum] = val
	}

	return callF(fd, params)
}

func callF(fd *funcDesc, params []reflect.Value) (interface{}, error) {
	for i, val := range params {
		if i != fd.ctxNum && val.Kind() == reflect.Ptr {
			params[i] = val.Elem()
		}
	}

	res := fd.f.Call(params)

	l := len(res)
	var result interface{}
	var err error
	for i := 0; i < l; i++ {
		v := res[i]
		if !v.IsValid() {
			continue
		}
		if v.Kind() == reflect.Interface || v.Kind() == reflect.Ptr {
			if v.IsNil() {
				continue
			}
		}
		if e, ok := v.Interface().(error); ok && err == nil {
			err = e
			continue
		}
		if result == nil { // && !v.IsNil()
			result = v.Interface()
			continue
		}
	}
	return result, err
}

func writeXmlBody(ctx *Ctx, env soap.SOAPEnvelope) {
	ctx.mappedResponseBody = env.Body.Content

	bytes, err := xml.Marshal(env)
	if err != nil {
		logger.Warn(err)
		ctx.SetStatusCode(http.StatusInternalServerError)
	}
	ctx.SetContentType(xmlContentType)
	ctx.Write(bytes)
}

func writeJsonBody(ctx *Ctx, result interface{}) {
	ctx.mappedResponseBody = result

	bytes, err := json.Marshal(result)
	if err != nil {
		logger.Warn(err)
		ctx.SetStatusCode(http.StatusInternalServerError)
	}
	ctx.SetContentType(jsonContentType)
	ctx.Write(bytes)
}

func getActionKey(ctx *fasthttp.RequestCtx) string {
	headerValue := string(ctx.Request.Header.Peek(soapActionHeader))
	unquoted, err := strconv.Unquote(headerValue)
	if err != nil {
		unquoted = headerValue
	}
	return toKey(string(ctx.Request.URI().RequestURI()), unquoted)
}

func toKey(uri, action string) string {
	if action == "" {
		return uri
	}
	return fmt.Sprintf("%s/%s", uri, action)
}
