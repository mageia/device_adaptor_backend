package logger

import (
	"io"
	"log"
	"os"
	"regexp"
	"time"
)

var prefixRegex = regexp.MustCompile("^[DIWE]!")

type Logger struct {
	writer io.Writer
}

func (l *Logger) Write(b []byte) (n int, err error) {
	var line []byte
	if !prefixRegex.Match(b) {
		line = append([]byte(time.Now().UTC().Format(time.RFC3339)+" I! "), b...)
	} else {
		line = append([]byte(time.Now().UTC().Format(time.RFC3339)+" "), b...)
	}
	return l.writer.Write(line)
}

func newLogger(w io.Writer) io.Writer {
	return &Logger{
		writer: NewWriter(w),
	}
}

func SetupLogging(debug bool, logfile string) {
	var oFile *os.File

	if debug {
		SetLevel(DEBUG)
		log.SetFlags(log.Lshortfile)
	}

	if logfile != "" {
		if _, err := os.Stat(logfile); os.IsNotExist(err) {
			if oFile, err = os.Create(logfile); err != nil {
				log.Printf("E! Unable to create %s (%s), using stderr", logfile, err)
				oFile = os.Stderr
			}
		} else {
			if oFile, err = os.OpenFile(logfile, os.O_APPEND|os.O_WRONLY, os.ModeAppend); err != nil {
				log.Printf("E! Unable to append to %s (%s), using stderr", logfile, err)
				oFile = os.Stderr
			}
		}
	} else {
		oFile = os.Stderr
	}

	log.SetOutput(newLogger(oFile))
}
