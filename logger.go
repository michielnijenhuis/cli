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

func defaultFormatter(message string, level string, ctx any) string {
	formattedTime := time.Now().Format("15:04:05")
	color := logLevelToColor(level)

	mj := ""
	if ctx != nil {
		j, err := json.Marshal(ctx)
		if err != nil {
			mj = Eol + Dim(string(j))
		}
	}

	return fmt.Sprintf("<fg=%s>[%s] %s: %s</>%s", color, formattedTime, strings.ToUpper(level), message, mj)
}

func iconFormatter(message string, level string, ctx any) string {
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

func boxFormatter(message string, level string, ctx any) string {
	formattedTime := Dim(time.Now().Format("15:04:05"))
	color := logLevelToColor(level)
	mj := ""
	if ctx != nil {
		j, err := json.Marshal(ctx)
		if err != nil {
			mj = Eol + Dim(string(j))
		}
	}
	return strings.TrimSuffix(Box(fmt.Sprintf("%s <fg=%s>%s</>", formattedTime, color, strings.ToUpper(level)), message+mj, "", "", ""), Eol)
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
	l.log(message, "debug", nil)
}

func (l *logger) Debugc(message string, ctx any) {
	l.log(message, "debug", ctx)
}

func (l *logger) Debugf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), "debug", nil)
}

func (l *logger) Info(message string) {
	l.log(message, "info", nil)
}

func (l *logger) Infoc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), "info", ctx)
}

func (l *logger) Infof(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), "info", nil)
}

func (l *logger) Notice(message string) {
	l.log(message, "notice", nil)
}

func (l *logger) Noticec(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), "notice", ctx)
}

func (l *logger) Noticef(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), "notice", nil)
}

func (l *logger) Warn(message string) {
	l.log(message, "warn", nil)
}

func (l *logger) Warnc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), "warn", ctx)
}

func (l *logger) Warnf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), "warn", nil)
}

func (l *logger) Error(message string) {
	l.log(message, "error", nil)
}

func (l *logger) Errorc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), "error", ctx)
}

func (l *logger) Errorf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), "error", nil)
}

func (l *logger) Critical(message string) {
	l.log(message, "critical", nil)
}

func (l *logger) Criticalc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), "critical", ctx)
}

func (l *logger) Criticalf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), "critical", nil)
}

func (l *logger) Alert(message string) {
	l.log(message, "alert", nil)
}

func (l *logger) Alertc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), "alert", ctx)
}

func (l *logger) Alertf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), "alert", nil)
}

func (l *logger) Fatal(message string) {
	l.log(message, "fatal", nil)
}

func (l *logger) Fatalc(format string, ctx any) {
	l.log(fmt.Sprintf(format, ctx), "fatal", ctx)
}

func (l *logger) Fatalf(format string, args ...any) {
	l.log(fmt.Sprintf(format, args...), "fatal", nil)
}

func (l *logger) log(message string, level string, ctx any) {
	var formatter logFormatter
	switch currentTheme {
	case ThemeBlock:
		formatter = boxFormatter
	case ThemeIcon:
		formatter = iconFormatter
	default:
		formatter = defaultFormatter
	}

	if l.debug || l.level <= levelToInt(level) {
		l.o.Writeln(formatter(message, level, ctx), 0)

		if level == "fatal" {
			os.Exit(1)
		}
	}
}

func levelToInt(level string) int {
	switch level {
	case "debug":
		return 0
	case "info":
		return 1
	case "notice":
		return 2
	case "warn":
		return 3
	case "error":
		return 4
	case "critical":
		return 5
	case "alert":
		return 6
	case "fatal":
		return 7
	default:
		return 0
	}
}

func levelToTheme(level string) *Theme {
	switch level {
	case "debug":
		return nil
	case "info", "notice":
		t, _ := GetTheme("info")
		return t
	case "warn":
		t, _ := GetTheme("warn")
		return t
	case "error", "critical", "alert", "fatal":
		t, _ := GetTheme("error")
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
	case "debug":
		return "white"
	case "info":
		return "bright-blue"
	case "notice":
		return "bright-blue"
	case "warn":
		return "bright-yellow"
	case "error":
		return "bright-red"
	case "critical":
		return "bright-red"
	case "alert":
		return "bright-red"
	case "fatal":
		return "bright-red"
	default:
		return "white"
	}
}
