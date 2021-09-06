package subprocess

import (
	"io"
	"strings"

	"github.com/ARM-software/golang-utils/utils/logs"
	"github.com/ARM-software/golang-utils/utils/platform"
)

var lineSep = platform.UnixLineSeparator()

type logStreamer struct {
	io.Writer
	IsStdErr bool
	Loggers  logs.Loggers
}

func (l *logStreamer) Write(p []byte) (n int, err error) {
	lines := strings.Split(string(p), lineSep)
	for i := range lines { // https://stackoverflow.com/questions/62446118/implicit-memory-aliasing-in-for-loop
		line := lines[i]
		if line != "" {
			if l.IsStdErr {
				l.Loggers.LogError(line)
			} else {
				l.Loggers.Log(line)
			}
		}
	}
	return len(p), nil
}

func newLogStreamer(IsStdErr bool, Loggers logs.Loggers) *logStreamer {
	return &logStreamer{
		IsStdErr: IsStdErr,
		Loggers:  Loggers,
	}
}

func newOutStreamer(Loggers logs.Loggers) *logStreamer {
	return newLogStreamer(false, Loggers)
}

func newErrLogStreamer(Loggers logs.Loggers) *logStreamer {
	return newLogStreamer(true, Loggers)
}
