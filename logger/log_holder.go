// Deprecated: use github.com/integration-system/isp-log instead
package logger

import (
	"fmt"
	"github.com/integration-system/isp-lib/utils"
	"github.com/sirupsen/logrus"
	"github.com/x-cray/logrus-prefixed-formatter"
	"os"
	"strings"
)

const (
	defaultTsFormat = "2006-01-02 15:04:05.000Z07:00"
	alertMsgFormat  = "============== %s =============="
)

var (
	terminalLogger *logrus.Logger
)

type log2LogrusWriter struct {
}

func (w *log2LogrusWriter) Write(b []byte) (int, error) {
	return os.Stdout.Write(b)
}

func init() {
	terminalLogger = logrus.New()
	terminalFormatter := &prefixed.TextFormatter{
		ForceFormatting:  true,
		DisableColors:    !utils.DEV,
		ForceColors:      utils.DEV,
		FullTimestamp:    true,
		DisableTimestamp: false,
		TimestampFormat:  defaultTsFormat,
		QuoteCharacter:   "",
	}
	terminalFormatter.SetColorScheme(&prefixed.ColorScheme{
		TimestampStyle:  "white",
		InfoLevelStyle:  "cyan+h",
		DebugLevelStyle: "cyan",
		WarnLevelStyle:  "yellow",
		ErrorLevelStyle: "red",
		PanicLevelStyle: "red",
		FatalLevelStyle: "red+b",
	})
	//hook, err := lSyslog.NewSyslogHook("", "", syslog.LOG_INFO, "")

	/*if err == nil {
		log.Println("logrus syslog hook created successfully", err)
		terminalLogger.AddHook(hook)
	}*/
	terminalLogger.Formatter = terminalFormatter
	terminalLogger.SetOutput(&log2LogrusWriter{})
	terminalLogger.AddHook(defaultHook())

	if utils.DEV {
		SetLevel(logrus.DebugLevel.String())
	} else {
		logLevel := utils.LOG_LEVEL
		if logLevel == "" {
			logLevel = "info"
		}
		SetLevel(logLevel)
	}
}

func SetLevel(level string) {
	if level, err := logrus.ParseLevel(level); err == nil {
		terminalLogger.SetLevel(level)
	} else {
		Fatal(err)
	}
}

// Deprecated: use github.com/integration-system/isp-log
func Info(args ...interface{}) {
	terminalLogger.Infoln(args...)

}

// Deprecated: use github.com/integration-system/isp-log
func Infof(format string, args ...interface{}) {
	terminalLogger.Infof(format, args...)

}

// Deprecated: use github.com/integration-system/isp-log
func Warn(args ...interface{}) {
	terminalLogger.Warnln(args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Warnf(format string, args ...interface{}) {
	terminalLogger.Warnf(format, args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Debug(args ...interface{}) {
	terminalLogger.Debugln(args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Debugf(format string, args ...interface{}) {
	terminalLogger.Debugf(format, args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Error(args ...interface{}) {
	terminalLogger.Errorln(args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Errorf(format string, args ...interface{}) {
	terminalLogger.Errorf(format, args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Fatal(args ...interface{}) {
	terminalLogger.Fatalln(args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Fatalf(format string, args ...interface{}) {
	terminalLogger.Fatalf(format, args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Panic(args ...interface{}) {
	terminalLogger.Panicln(args)
}

// Deprecated: use github.com/integration-system/isp-log
func Panicf(format string, args ...interface{}) {
	terminalLogger.Panicf(format, args)
}

// Deprecated: use github.com/integration-system/isp-log
func Log(level logrus.Level, args ...interface{}) {
	terminalLogger.Log(level, args...)
}

// Deprecated: use github.com/integration-system/isp-log
func Logf(level logrus.Level, format string, args ...interface{}) {
	terminalLogger.Logf(level, format, args...)
}

func FmtAlertMsg(msg string) string {
	return fmt.Sprintf(alertMsgFormat, strings.ToUpper(msg))
}
