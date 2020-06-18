package clog

import (
	"fmt"
	"io"
	"time"

	"github.com/castisdev/cilog"
)

// Set :
func Set(dir, module, moduleVersion string, minLevel cilog.Level) {
	cilog.Set(cilog.NewLogWriter(dir, module, 10*1024*1024), module, moduleVersion, minLevel)
}

// SetWriter :
func SetWriter(w io.Writer) {
	cilog.SetWriter(w)
}

// SetMinLevel :
func SetMinLevel(lvl cilog.Level) {
	cilog.SetMinLevel(lvl)
}

// SetDir :
func SetDir(dir string) {
	cilog.SetWriter(cilog.NewLogWriter(dir, cilog.GetModule(), 10*1024*1024))
}

// Debugf1 :
func Debugf1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.DEBUG, msg, time.Now())
}

// Reportf1 :
func Reportf1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.REPORT, msg, time.Now())
}

// Infof1 :
func Infof1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.INFO, msg, time.Now())
}

// Successf1 :
func Successf1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.SUCCESS, msg, time.Now())
}

// Warningf1 :
func Warningf1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.WARNING, msg, time.Now())
}

// Errorf1 :
func Errorf1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.ERROR, msg, time.Now())
}

// Failf1 :
func Failf1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.FAIL, msg, time.Now())
}

// Exceptionf1 :
func Exceptionf1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.EXCEPTION, msg, time.Now())
}

// Criticalf1 :
func Criticalf1(traceID string, format string, v ...interface{}) {
	msg := "[" + traceID + "] " + fmt.Sprintf(format, v...)
	cilog.StdLogger().Log(2, cilog.CRITICAL, msg, time.Now())
}

// Debugf :
func Debugf(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.DEBUG, fmt.Sprintf(format, v...), time.Now())
}

// Reportf :
func Reportf(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.REPORT, fmt.Sprintf(format, v...), time.Now())
}

// Infof :
func Infof(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.INFO, fmt.Sprintf(format, v...), time.Now())
}

// Successf :
func Successf(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.SUCCESS, fmt.Sprintf(format, v...), time.Now())
}

// Warningf :
func Warningf(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.WARNING, fmt.Sprintf(format, v...), time.Now())
}

// Errorf :
func Errorf(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.ERROR, fmt.Sprintf(format, v...), time.Now())
}

// Failf :
func Failf(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.FAIL, fmt.Sprintf(format, v...), time.Now())
}

// Exceptionf :
func Exceptionf(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.EXCEPTION, fmt.Sprintf(format, v...), time.Now())
}

// Criticalf :
func Criticalf(format string, v ...interface{}) {
	cilog.StdLogger().Log(2, cilog.CRITICAL, fmt.Sprintf(format, v...), time.Now())
}

// IsDebugEnable :
func IsDebugEnable() bool {
	return cilog.GetMinLevel() == cilog.DEBUG
}
