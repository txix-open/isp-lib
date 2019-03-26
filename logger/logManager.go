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
	fileLogger     *logrus.Logger
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
		if fileLogger != nil {
			fileLogger.SetLevel(level)
		}
	} else {
		Fatal(err)
	}
}

func Info(args ...interface{}) {
	terminalLogger.Infoln(args...)
	if fileLogger != nil {
		fileLogger.Infoln(args...)
	}
}

func Infof(format string, args ...interface{}) {
	terminalLogger.Infof(format, args...)
	if fileLogger != nil {
		fileLogger.Infof(format, args...)
	}
}

func Warn(args ...interface{}) {
	terminalLogger.Warnln(args...)
	if fileLogger != nil {
		fileLogger.Warnln(args...)
	}
}

func Warnf(format string, args ...interface{}) {
	terminalLogger.Warnf(format, args...)
	if fileLogger != nil {
		fileLogger.Warnf(format, args...)
	}
}

func Debug(args ...interface{}) {
	terminalLogger.Debugln(args...)
	if fileLogger != nil {
		fileLogger.Debugln(args...)
	}
}

func Debugf(format string, args ...interface{}) {
	terminalLogger.Debugf(format, args...)
	if fileLogger != nil {
		fileLogger.Debugf(format, args...)
	}
}

func Error(args ...interface{}) {
	terminalLogger.Errorln(args...)
	if fileLogger != nil {
		fileLogger.Errorln(args...)
	}
}

func Errorf(format string, args ...interface{}) {
	terminalLogger.Errorf(format, args...)
	if fileLogger != nil {
		fileLogger.Errorf(format, args...)
	}
}

func Fatal(args ...interface{}) {
	if fileLogger != nil {
		fileLogger.Errorln(args...)
	}
	terminalLogger.Fatalln(args...)
}

func Fatalf(format string, args ...interface{}) {
	if fileLogger != nil {
		fileLogger.Errorf(format, args...)
	}
	terminalLogger.Fatalf(format, args...)
}

func Panic(args ...interface{}) {
	if fileLogger != nil {
		fileLogger.Errorln(args...)
	}
	terminalLogger.Panicln(args)
}

func Panicf(format string, args ...interface{}) {
	if fileLogger != nil {
		fileLogger.Errorf(format, args...)
	}
	terminalLogger.Panicf(format, args)
}

func FmtAlertMsg(msg string) string {
	return fmt.Sprintf(alertMsgFormat, strings.ToUpper(msg))
}
