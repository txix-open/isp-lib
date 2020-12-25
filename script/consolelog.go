package script

import (
	"bytes"
	"encoding/json"
)

type consoleLog struct {
	logBuf *bytes.Buffer
}

func (cl *consoleLog) log(args ...interface{}) {
	tmp, err := json.Marshal(args)
	if err != nil {
		panic(err)
	}
	cl.logBuf.Write(tmp)
	cl.logBuf.Write([]byte(",\n"))
}

func newConsoleLog(logBuf *bytes.Buffer) map[string]interface{} {
	if logBuf == nil {
		return map[string]interface{}{
			"log": func(args ...interface{}) {},
		}
	}
	newConsoleLog := &consoleLog{logBuf: logBuf}
	return map[string]interface{}{
		"log": newConsoleLog.log,
	}
}
