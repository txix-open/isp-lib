package utils

import (
	"fmt"
	"strings"
	"time"
)

type JSONTime struct {
	time.Time
}

func (t JSONTime) MarshalJSON() ([]byte, error) {
	// do your serializing here
	stamp := fmt.Sprintf("\"%s\"", time.Time(t.Time).Format(FullDateFormat))
	return []byte(stamp), nil
}

func (t *JSONTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		t.Time = time.Time{}
		return
	}
	t.Time, err = time.Parse(FullDateFormat, s)
	return
}
