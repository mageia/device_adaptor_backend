package logger

import (
	"github.com/sirupsen/logrus"
	"log"
)

var Logger = logrus.New()

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{ForceColors: true})
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
