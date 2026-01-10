package logger

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func Init(verbose bool) {
	log = logrus.New()

	if verbose {
		log.SetLevel(logrus.DebugLevel)
		log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	} else {
		log.SetLevel(logrus.InfoLevel)
		log.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
		})
	}

	// Output to stdout
	log.SetOutput(os.Stdout)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return log.WithFields(fields)
}

func Info(msg string) {
	log.Info(msg)
}

func Infof(format string, args ...interface{}) {
	log.Infof(format, args...)
}

func Debug(msg string) {
	log.Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	log.Debugf(format, args...)
}

func Warn(msg string) {
	log.Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	log.Warnf(format, args...)
}

func Error(msg string) {
	log.Error(msg)
}

func Errorf(format string, args ...interface{}) {
	log.Errorf(format, args...)
}

func Fatal(msg string) {
	log.Fatal(msg)
}

func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

func SetOutput(output io.Writer) {
	log.SetOutput(output)
}

func GetLogger() *logrus.Logger {
	return log
}
