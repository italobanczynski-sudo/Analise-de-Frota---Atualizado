// Package log configura o logrus usado por todo o servidor.
package log

import (
	"os"

	"github.com/sirupsen/logrus"
)

func Init() {
	logrus.SetOutput(os.Stdout)
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05"})
	logrus.SetLevel(logrus.InfoLevel)
}
