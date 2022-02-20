package logger

import (
	"context"
	"fmt"
	nested "github.com/antonfisher/nested-logrus-formatter"
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"
)

const (
	// Silent silent log level
	Silent LogLevel = iota + 1
	// Error error log level
	Error
	// Warn warn log level
	Warn
	// Info info log level
	Info
	Debug
	Trace

	DefaultLogPath       = "./logs"
	DefaultLogFile       = "eagle.log"
	DefaultRotationCount = uint(30)
	DefaultRotationTime  = time.Duration(24) * time.Hour
)

/*
	注意：此为业务型日志，项目启动初始化时建议还是使用系统日志会好点，减少日志文件生成
		例如， print和fatal这种涉及系统级别的操作，使用原生会好点，没必要记录
		业务日志一般记录等级都有 trace，debug, info, warning, error级别即可
*/

type (
	Interface interface {
		LogMode(LogLevel)
		Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error)
		Debug(context.Context, string, ...interface{})
		Info(context.Context, string, ...interface{})
		Warn(context.Context, string, ...interface{})
		Error(context.Context, string, ...interface{})
		//SetFormat(format *nested.Formatter)
		//WithField(fields logrus.Fields)
	}

	Config struct {
		Ctx           context.Context
		LogLevel      LogLevel
		CreateFile    bool
		LogPath       string
		LogFile       string
		RotationCount uint
		RotationTime  time.Duration
	}

	// LogLevel log level
	LogLevel int

	// log
	fenvLogger struct {
		Config
		writer
	}

	writer struct {
		instance *logrus.Logger
	}
)

func New(c Config) Interface {
	// 初始化logrus日志对象
	logrusInstance := logrus.New()
	logrusInstance.SetOutput(os.Stdout)
	logrusInstance.SetReportCaller(false)
	if c.CreateFile {
		if len(c.LogFile) == 0 {
			c.LogFile = DefaultLogFile
		}
		if len(c.LogPath) == 0 {
			c.LogPath = DefaultLogPath
		}
		logFileName := path.Join(c.LogPath, c.LogFile)
		rotationCount := DefaultRotationCount
		rotationTime := DefaultRotationTime
		if c.RotationCount > 0 {
			rotationCount = c.RotationCount
		}
		if c.RotationTime > 0 {
			rotationTime = c.RotationTime
		}
		// 使用滚动压缩方式记录日志
		w, _ := rotateLogs.New(
			logFileName+".%Y%m%d%H%M",
			rotateLogs.WithLinkName(logFileName),
			rotateLogs.WithRotationCount(rotationCount),
			rotateLogs.WithRotationTime(rotationTime),
		)
		mv := io.MultiWriter(os.Stdout, w)
		logrusInstance.SetOutput(mv)
		logrusInstance.SetFormatter(
			&nested.Formatter{
				HideKeys:        true,
				NoColors:        true,
				TimestampFormat: "2006-01-02 15:04:05",
				CallerFirst:     true,
				CustomCallerFormatter: func(frame *runtime.Frame) string {
					funcInfo := runtime.FuncForPC(frame.PC)
					if funcInfo == nil {
						return "error during runtime.FuncForPC"
					}
					fullPath, line := funcInfo.FileLine(frame.PC)
					return fmt.Sprintf(" [%v:%v] ", filepath.Base(fullPath), line)
				},
			})
	}
	w := writer{logrusInstance}
	fl := &fenvLogger{
		Config: c,
		writer: w,
	}
	if c.LogLevel == 0 {
		fl.LogMode(Debug)
	} else {
		fl.LogMode(fl.LogLevel)
	}
	return fl
}

// LogMode log mode
func (f *fenvLogger) LogMode(level LogLevel) {
	f.LogLevel = level
	switch level {
	case Debug:
		f.instance.SetReportCaller(true)
		f.instance.SetLevel(logrus.DebugLevel)
	case Info:
		f.instance.SetLevel(logrus.InfoLevel)
	case Warn:
		f.instance.SetLevel(logrus.WarnLevel)
	case Error:
		f.instance.SetLevel(logrus.ErrorLevel)
	case Trace:
		f.instance.SetLevel(logrus.TraceLevel)
	default:
		f.instance.SetLevel(logrus.InfoLevel)
	}
}

func (f fenvLogger) Debug(ctx context.Context, msg string, data ...interface{}) {
	if f.LogLevel >= Debug {
		f.print(f.LogLevel, msg, data)
	}
}

// Info print info
func (f fenvLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if f.LogLevel >= Info {
		f.print(f.LogLevel, msg, data)
	}
}

// Warn print warn messages
func (f fenvLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if f.LogLevel >= Warn {
		f.print(f.LogLevel, msg, data)
	}
}

// Error print error messages
func (f fenvLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if f.LogLevel >= Error {
		f.print(f.LogLevel, msg, data)
	}
}

// Trace print sql message
func (f fenvLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if f.LogLevel >= Trace {
		// trace 单独处理
	}
}

// 后续再优化：https://darjun.github.io/2020/02/07/godailylib/logrus/ 自定义属性格式
/*func (f fenvLogger)WithField(field logrus.Fields) {
	f.instance.WithFields(field)
}*/
// 参考 https://www.codeleading.com/article/82454852589/
/*func (f fenvLogger)SetFormat(format *nested.Formatter){
	f.instance.SetFormatter(format)
}*/

func (w *writer) print(level LogLevel, msg string, args ...interface{}) {
	switch level {
	case Debug:
		w.instance.Debugf(msg, args)
	case Info:
		w.instance.Infof(msg, args)
	case Warn:
		w.instance.Warnf(msg, args)
	case Error:
		w.instance.Errorf(msg, args)
	case Trace:
		w.instance.Tracef(msg, args)
	default:
		w.instance.Print(msg, args)
	}
}

/*
	------ 以下是涉及系统启动或者崩溃的日志记录，不会存入文件 ------
*/

func Printf(format string, v ...interface{}) {
	log.Printf(format, v)
}

func Println(msg ...interface{}) {
	log.Println(msg)
}

func Fatal(msg ...interface{}) {
	log.Fatal(msg)
}
