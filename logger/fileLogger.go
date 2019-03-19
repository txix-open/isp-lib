package logger

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"github.com/integration-system/isp-lib/structure"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

const timestampFormat = "2006-01-02T15:04:05.999-07:00"

type SyncLogger interface {
	Log(event, source string, data string) error
	io.Closer
}

func NewFileLogger(config structure.SyncLoggerConfig) (SyncLogger, error) {
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
		var fw *bufio.Writer
		var gf *gzip.Writer
		if config.Compress {
			gf = gzip.NewWriter(file)
			fw = bufio.NewWriter(gf)
		} else {
			fw = bufio.NewWriter(file)
		}

		return &logger{fw: fw, gf: gf, log: file, compress: config.Compress, immediateFlush: config.ImmediateFlush}, nil
	}
}

type logger struct {
	log            *os.File
	lock           sync.Mutex
	gf             *gzip.Writer
	fw             *bufio.Writer
	compress       bool
	immediateFlush bool
}

func (dfl *logger) Log(event, source string, data string) error {
	dfl.lock.Lock()
	_, err := dfl.fw.WriteString(time.Now().Format(timestampFormat) + " " + event + " " + source + " " + data + "\n\n")
	if dfl.immediateFlush {
		dfl.fw.Flush()
		if dfl.compress {
			err = dfl.gf.Flush()
		}
	}
	dfl.lock.Unlock()
	return err
}

func (dfl *logger) Close() error {
	dfl.close()
	return nil
}

func (dfl *logger) close() {
	err := dfl.fw.Flush()
	if err != nil {
		Warn("An error occurred when were tried to flush buffer data into a log file", err)
	}

	if dfl.compress {
		err = dfl.gf.Flush()
		if err != nil {
			Warn("An error occurred when were tried to flush gz data into a log file", err)
		}

		// Close the gzip first.
		err = dfl.gf.Close()
		if err != nil {
			Warn("An error occurred when were tried to close gzip writer", err)
		}
	}

	err = dfl.log.Close()
	if err != nil {
		Warn("An error occurred when were tried to close file", err)
	}

	Info(FmtAlertMsg("LOG FILE CLOSED"))
}
