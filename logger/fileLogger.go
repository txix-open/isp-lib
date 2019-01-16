package logger

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

const timestampFormat = "2006-01-02T15:04:05.999-07:00"

type SyncLoggerConfig struct {
	Enable     bool   `schema:"Enable file logging"`
	Filename   string `json:"filename" yaml:"filename" schema:"File name"`
	MaxSize    int    `json:"-" yaml:"maxsize"`
	MaxAge     int    `json:"-" yaml:"maxage"`
	MaxBackups int    `json:"-" yaml:"maxbackups"`
	LocalTime  bool   `json:"-" yaml:"localtime"`
	Compress   bool   `json:"-" yaml:"compress"`
}

type SyncLogger interface {
	Log(event, source string, data string) error
	io.Closer
}

func NewGZipLogger(config SyncLoggerConfig) (SyncLogger, error) {
	if !config.Enable {
		return nil, nil
	} else {
		dir, _ := path.Split(config.Filename)
		if dir != "" {
			err := os.MkdirAll(dir, os.ModeDir|0775)
			if err != nil {
				return nil, fmt.Errorf(
					"An error occurred when were tried to create the directory: %s for log files: %v",
					dir,
					err)
			}
		}

		file, err := os.OpenFile(config.Filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0775)
		if err != nil {
			return nil, fmt.Errorf(
				"An error occurred when were tried to open the log file: %s : %v",
				config.Filename,
				err)
		}
		gf := gzip.NewWriter(file)
		fw := bufio.NewWriter(gf)

		return &gzipLogger{fw: fw, gf: gf, log: file}, nil
	}
}

type gzipLogger struct {
	log  *os.File
	lock sync.Mutex
	gf   *gzip.Writer
	fw   *bufio.Writer
}

func (dfl *gzipLogger) Log(event, source string, data string) error {
	dfl.lock.Lock()
	_, err := dfl.fw.WriteString(time.Now().Format(timestampFormat) + " " + event + " " + source + " " + data + "\n")
	dfl.lock.Unlock()
	return err
}

func (dfl *gzipLogger) Close() error {
	dfl.closeGZ()
	return nil
}

func (dfl *gzipLogger) closeGZ() {
	err := dfl.fw.Flush()
	if err != nil {
		Warn("An error occurred when were tried to flush buffer data into a log file", err)
	}

	err = dfl.gf.Flush()
	if err != nil {
		Warn("An error occurred when were tried to flush gz data into a log file", err)
	}

	// Close the gzip first.
	err = dfl.gf.Close()
	if err != nil {
		Warn("An error occurred when were tried to close gzip writer", err)
	}

	err = dfl.log.Close()
	if err != nil {
		Warn("An error occurred when were tried to close file", err)
	}

	Info(FmtAlertMsg("LOG FILE CLOSED"))
}
