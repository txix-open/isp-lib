package utils

import (
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
	"github.com/modern-go/reflect2"
	"time"
	"unicode"
	"unsafe"
)

const (
	empty = ""
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

func ConvertBytesToGo(b []byte, ptr interface{}) error {
	return ji.Unmarshal(b, ptr)
}

func ConvertGoToBytes(data interface{}) ([]byte, error) {
	return ji.Marshal(data)
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
