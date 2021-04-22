package logger

import (
	"io"
	"io/ioutil"
	"log"
	"os"

	"github.com/rs/zerolog"
)

type WriterFilter struct {
	w   io.Writer
	lvl zerolog.Level
}

func (w *WriterFilter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}

func NewWriterFilter(w io.Writer, lvl zerolog.Level) *WriterFilter {
	return &WriterFilter{
		w,
		lvl,
	}
}

func (w *WriterFilter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= w.lvl {
		return w.w.Write(p)
	}
	return len(p), nil
}

func Level(lvl string) zerolog.Level {
	level := zerolog.InfoLevel

	switch lvl {
	case "Debug":
		level = zerolog.DebugLevel
	case "Warning":
		level = zerolog.WarnLevel
	case "Error":
		level = zerolog.ErrorLevel
	case "Info":
		level = zerolog.InfoLevel
	}

	return level
}

func LevelN(n int) zerolog.Level {
	var level zerolog.Level

	if n >= 4 {
		level = zerolog.DebugLevel
	} else if n < 0 {
		level = zerolog.Disabled
	} else {
		switch n {
		case 3:
			level = zerolog.InfoLevel
		case 2:
			level = zerolog.WarnLevel
		case 1:
			level = zerolog.ErrorLevel
		case 0:
			level = zerolog.FatalLevel
		}
	}

	return level
}

func Writer(path string) io.Writer {
	var writer io.Writer

	switch path {
	case "none":
		writer = ioutil.Discard
	default:
		file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
		if err != nil {
			log.Fatalln("Cannot open log file", err)
		}
		writer = file
	}

	return writer
}
