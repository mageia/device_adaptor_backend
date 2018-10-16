package logger

import (
	"io"
	"sync"
)

type Level int

const (
	_ Level = iota
	DEBUG
	INFO
	WARN
	ERROR
	OFF
)

const Delimiter = '!'

var invalidMSG = []byte("log messages must have 'L!' prefix where L is one of 'D', 'I', 'W', 'E'")

var logLevel = INFO

var Levels = map[byte]Level{
	'D': DEBUG,
	'I': INFO,
	'W': WARN,
	'E': ERROR,
}

var mu sync.RWMutex

func SetLevel(l Level) {
	mu.Lock()
	defer mu.Unlock()
	logLevel = l
}

func LogLevel() Level {
	mu.RLock()
	defer mu.RUnlock()
	return logLevel
}

type Writer struct {
	start int
	w     io.Writer
}

func (w *Writer) Write(buf []byte) (int, error) {
	if len(buf) > 0 {
		if w.start == -1 {
			for i, c := range buf {
				if c == Delimiter && i > 0 {
					if Levels[buf[i-1]] > 0 {
						w.start = i - 1
						break
					}
				}
			}
			if w.start == -1 {
				buf = append(invalidMSG, buf...)
				return w.w.Write(buf)
			}
		}

		l := Levels[buf[w.start]]
		if l >= LogLevel() {
			return w.w.Write(buf)
		} else if l == 0 {
			buf = append(invalidMSG, buf...)
			return w.w.Write(buf)
		}
	}
	return 0, nil
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{-1, w}
}

//func New(w io.Writer, prefix string, flag int) *log.Logger {
//	return log.New(NewWriter(w), prefix, flag)
//}
