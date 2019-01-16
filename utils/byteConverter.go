package utils

import (
	"bytes"
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/modern-go/reflect2"
	"golang.org/x/net/context"
	"google.golang.org/grpc/stats"
	"sync"
	"time"
	"unicode"
	"unsafe"
)

const (
	bufKey = "buf"
	empty  = ""
)

var (
	ji = jsoniter.ConfigFastest
)

func init() {
	extra.SetNamingStrategy(toCamelCase)

	tc := &timeCoder{}
	timeType := reflect2.TypeByName("time.Time")

	encExt := jsoniter.EncoderExtension{timeType: tc}
	decExt := jsoniter.DecoderExtension{timeType: tc}
	ji.RegisterExtension(encExt)
	ji.RegisterExtension(decExt)
}

type bufInjector struct {
	stats.Handler
	pool *sync.Pool
}

func (bi *bufInjector) TagRPC(ctx context.Context, i *stats.RPCTagInfo) context.Context {
	buf := bi.pool.Get().(*bytes.Buffer)
	ctx = context.WithValue(ctx, bufKey, buf)
	return ctx
}

func (bi *bufInjector) HandleRPC(ctx context.Context, s stats.RPCStats) {
	if _, ok := s.(*stats.End); ok {
		buf, ok := ctx.Value(bufKey).(*bytes.Buffer)
		if ok {
			buf.Reset()
			bi.pool.Put(buf)
		}
	}
}

func (bi *bufInjector) TagConn(ctx context.Context, s *stats.ConnTagInfo) context.Context {
	return ctx
}

func (bi *bufInjector) HandleConn(ctx context.Context, s stats.ConnStats) {

}

func ConvertBytesToGo(b []byte, ptr interface{}) error {
	return ji.Unmarshal(b, ptr)
}

func ConvertInterfaceToBytes(data interface{}, ctx context.Context) ([]byte, error) {
	return ji.Marshal(data)
}

func NewBufferInjector() *bufInjector {
	return &bufInjector{
		pool: &sync.Pool{
			New: func() interface{} { return newBuf() },
		},
	}
}

func newBuf() *bytes.Buffer {
	arr := make([]byte, 0, 32*1024)
	return bytes.NewBuffer(arr)
}

func toCamelCase(s string) string {
	if s == empty {
		return s
	}
	arr := []rune(s)
	arr[0] = unicode.ToLower(arr[0])
	return string(arr)
}

type timeCoder struct {
}

func (codec *timeCoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	t, err := time.Parse(FullDateFormat, iter.ReadString())
	if err != nil {
		iter.ReportError("string -> time.Time", err.Error())
	} else {
		*((*time.Time)(ptr)) = t
	}
}

func (codec *timeCoder) IsEmpty(ptr unsafe.Pointer) bool {
	ts := *((*time.Time)(ptr))
	return ts.IsZero()
}

func (codec *timeCoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	ts := *((*time.Time)(ptr))
	stream.WriteString(ts.Format(FullDateFormat))
}
