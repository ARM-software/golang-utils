package logs

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type SlowWriter struct {
	StdWriter
}

func (w *SlowWriter) Write(p []byte) (n int, err error) {
	time.Sleep(10 * time.Millisecond)
	return os.Stdout.Write(p)
}

// Creates a logger to standard output/error
func CreateMultipleWriterLogger(prefix string) (loggers Loggers, err error) {
	writer, err := CreateMultipleWritersWithSource(&StdWriter{}, &SlowWriter{})
	loggers = &GenericLoggers{
		Output: log.New(writer, "["+prefix+"] Output: ", log.LstdFlags),
		Error:  log.New(writer, "["+prefix+"] Error: ", log.LstdFlags),
	}
	return
}

func TestMultipleWriters(t *testing.T) {
	stdloggers, err := CreateMultipleWriterLogger("Test")
	require.Nil(t, err)
	_testLog(t, stdloggers)
	time.Sleep(100 * time.Millisecond)
}
