package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Logger interface {
	Debug(message string)
	Debugc(message string, ctx any)
	Debugf(format string, args ...any)
	Info(message string)
	Infoc(message string, ctx any)
	Infof(format string, args ...any)
	Notice(message string)
	Noticec(message string, ctx any)
	Noticef(format string, args ...any)
	Warn(message string)
	Warnc(message string, ctx any)
	Warnf(format string, args ...any)
	Error(message string)
	Errorc(message string, ctx any)
	Errorf(format string, args ...any)
	Critical(message string)
	Criticalc(message string, ctx any)
	Criticalf(format string, args ...any)
	Alert(message string)
	Alertc(message string, ctx any)
	Alertf(format string, args ...any)
	Fatal(message string)
	Fatalc(message string, ctx any)
	Fatalf(format string, args ...any)
	EnableDebug()
	SetLogLevel(level int)
}

type logFormatter func(message string, level string, ctx any) string

type logger struct {
	level int
	o     *Output
	debug bool
}

const (
	LogLevelDebug = iota
	LogLevelInfo
	LogLevelNotice
	LogLevelWarn
	LogLevelError
	LogLevelCritical
	LogLevelAlert
	LogLevelFatal
)

const (
	LogTagDebug    = "debug"
	LogTagInfo     = "info"
	LogTagNotice   = "notice"
	LogTagWarn     = "warn"
	LogTagError    = "error"
	LogTagCritical = "critical"
	LogTagAlert    = "alert"
	LogTagFatal    = "fatal"
)

func formatter(message string, level string, ctx any) string {
	formattedTime := Dim("[" + time.Now().Format("15:04:05") + "]")
	theme := levelToTheme(level)
	color := logLevelToColor(level)

	mj := ""
	if ctx != nil {
		j, err := json.Marshal(ctx)
		if err != nil {
			mj = Eol + Dim(string(j))
		}
	}

	icon := ""
	if theme.Icon != "" {
		icon = theme.Icon + " "
	}

	return fmt.Sprintf("%s <fg=%s>%s%s:</> %s%s", formattedTime, color, icon, strings.ToUpper(level), message, mj)
}

func NewLogger(o *Output, level int) Logger {
	l := &logger{
		o:     o,
		debug: false,
	}

	l.SetLogLevel(level)

	return l
}

func (l *logger) EnableDebug() {
	l.debug = true
}

func (l *logger) Debug(message string) {
	l.log(message, LogTagDebug, nil)
}

func (l *logger) Debugc(message string, ctx any) {
	l.log(message, LogTagDebug, ctx)
}

func (l *logger) Debugf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), LogTagDebug, nil)
}

func (l *logger) Info(message string) {
	l.log(message, LogTagInfo, nil)
}

func (l *logger) Infoc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), LogTagInfo, ctx)
}

func (l *logger) Infof(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), LogTagInfo, nil)
}

func (l *logger) Notice(message string) {
	l.log(message, LogTagNotice, nil)
}

func (l *logger) Noticec(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), LogTagNotice, ctx)
}

func (l *logger) Noticef(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), LogTagNotice, nil)
}

func (l *logger) Warn(message string) {
	l.log(message, LogTagWarn, nil)
}

func (l *logger) Warnc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), LogTagWarn, ctx)
}

func (l *logger) Warnf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), LogTagWarn, nil)
}

func (l *logger) Error(message string) {
	l.log(message, LogTagError, nil)
}

func (l *logger) Errorc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), LogTagError, ctx)
}

func (l *logger) Errorf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), LogTagError, nil)
}

func (l *logger) Critical(message string) {
	l.log(message, LogTagCritical, nil)
}

func (l *logger) Criticalc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), LogTagCritical, ctx)
}

func (l *logger) Criticalf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), LogTagCritical, nil)
}

func (l *logger) Alert(message string) {
	l.log(message, LogTagAlert, nil)
}

func (l *logger) Alertc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), LogTagAlert, ctx)
}

func (l *logger) Alertf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), LogTagAlert, nil)
}

func (l *logger) Fatal(message string) {
	l.log(message, LogTagFatal, nil)
}

func (l *logger) Fatalc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), LogTagFatal, ctx)
}

func (l *logger) Fatalf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), LogTagFatal, nil)
}

func (l *logger) log(message string, level string, ctx any) {
	var formatter logFormatter = formatter
	theme, err := GetTheme(level)
	if theme != nil && err != nil && theme.LogFormatter != nil {
		formatter = theme.LogFormatter
	}

	if l.debug || l.level <= levelToInt(level) {
		l.o.Writeln(formatter(message, level, ctx), 0)

		if level == LogTagFatal {
			os.Exit(1)
		}
	}
}

func levelToInt(level string) int {
	switch level {
	case LogTagDebug:
		return 0
	case LogTagInfo:
		return 1
	case LogTagNotice:
		return 2
	case LogTagWarn:
		return 3
	case LogTagError:
		return 4
	case LogTagCritical:
		return 5
	case LogTagAlert:
		return 6
	case LogTagFatal:
		return 7
	default:
		return 0
	}
}

func levelToTheme(level string) *Theme {
	switch level {
	case LogTagDebug:
		return nil
	case LogTagInfo, LogTagNotice:
		t, _ := GetTheme(LogTagInfo)
		return t
	case LogTagWarn:
		t, _ := GetTheme(LogTagWarn)
		return t
	case LogTagError, LogTagCritical, LogTagAlert, LogTagFatal:
		t, _ := GetTheme(LogTagError)
		return t
	default:
		return nil
	}
}

func (l *logger) SetLogLevel(level int) {
	if level < LogLevelDebug {
		level = LogLevelError
	} else if level > LogLevelFatal {
		level = LogLevelFatal
	}

	l.level = level
}

func logLevelToColor(level string) string {
	switch level {
	case LogTagDebug:
		return ColorWhite
	case LogTagInfo:
		return ColorBrightBlue
	case LogTagNotice:
		return ColorBrightBlue
	case LogTagWarn:
		return ColorBrightYellow
	case LogTagError:
		return ColorBrightRed
	case LogTagCritical:
		return ColorBrightRed
	case LogTagAlert:
		return ColorBrightRed
	case LogTagFatal:
		return ColorBrightRed
	default:
		return ColorWhite
	}
}
