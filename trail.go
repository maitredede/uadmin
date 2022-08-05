package uadmin

import (
	"fmt"

	"github.com/maitredede/uadmin/colors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Reporting Levels
const (
	DEBUG     = 0
	WORKING   = 1
	INFO      = 2
	OK        = 3
	WARNING   = 4
	ERROR     = 5
	CRITICAL  = 6
	ALERT     = 7
	EMERGENCY = 8
)

var trailTag = map[int]string{
	DEBUG:     colors.Debug,
	WORKING:   colors.Working,
	INFO:      colors.Info,
	OK:        colors.OK,
	WARNING:   colors.Warning,
	ERROR:     colors.Error,
	CRITICAL:  colors.Critical,
	ALERT:     colors.Alert,
	EMERGENCY: colors.Emergency,
}

var levelMap = map[int]string{
	DEBUG:     "[  DEBUG ]   ",
	WORKING:   "[ WORKING]   ",
	INFO:      "[  INFO  ]   ",
	OK:        "[   OK   ]   ",
	WARNING:   "[ WARNING]   ",
	ERROR:     "[  ERROR ]   ",
	CRITICAL:  "[CRITICAL]   ",
	ALERT:     "[  ALERT ]   ",
	EMERGENCY: "[  EMERG ]   ",
}

var (
	log *zap.Logger
)

func init() {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	cfg.Encoding = "console"
	cfg.DisableStacktrace = true
	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	log = l
}

var zapMap = map[int]zapcore.Level{
	DEBUG:     zapcore.DebugLevel,
	WORKING:   zapcore.DebugLevel,
	INFO:      zapcore.InfoLevel,
	OK:        zapcore.InfoLevel,
	WARNING:   zapcore.WarnLevel,
	ERROR:     zapcore.ErrorLevel,
	CRITICAL:  zapcore.ErrorLevel,
	ALERT:     zapcore.WarnLevel,
	EMERGENCY: zapcore.ErrorLevel,
}

// Trail prints to the log
func Trail(level int, msg interface{}, i ...interface{}) {
	if level >= ReportingLevel {
		zl := zapMap[level]
		if ce := log.Check(zl, fmt.Sprintf(fmt.Sprint(msg), i...)); ce != nil {
			ce.Write()
		}
		// message := fmt.Sprint(msg)
		// if level != WORKING && !strings.HasSuffix(message, "\n") {
		// 	message += "\n"
		// } else if level == WORKING && !strings.HasPrefix(message, "\r") {
		// 	message = message + "\r"
		// }
		// if ReportTimeStamp {
		// 	log.Printf(trailTag[level]+message, i...)
		// } else {
		// 	fmt.Printf(trailTag[level]+message, i...)
		// }

		// // Run error handler if it exists
		// if ErrorHandleFunc != nil {
		// 	stack := string(debug.Stack())
		// 	stackList := strings.Split(stack, "\n")
		// 	stack = strings.Join(stackList[5:], "\n")
		// 	go ErrorHandleFunc(level, fmt.Sprintf(fmt.Sprint(msg), i...), stack)
		// }

		// // Log to syslog
		// if LogTrail && level >= TrailLoggingLevel && level != WORKING {
		// 	// Send log to syslog
		// 	Syslogf(level, message, i...)
		// }
	}
}
