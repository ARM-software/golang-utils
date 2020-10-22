package logs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLog(t *testing.T) {
	var loggers Loggers = &GenericLoggers{}
	err := loggers.Check()
	assert.NotNil(t, err)
	err = loggers.Close()
	assert.Nil(t, err)
}

func _testLog(t *testing.T, loggers Loggers) {
	err := loggers.Check()
	require.Nil(t, err)
	defer func() { _ = loggers.Close() }()

	err = loggers.SetLogSource("source1")
	require.Nil(t, err)
	err = loggers.SetLoggerSource("LoggerSource1")
	require.Nil(t, err)

	loggers.Log("Test output1")
	loggers.Log("Test output2")
	loggers.Log("\"/usr/bin/armlink\" --via=\"/workspace/Objects/aws_mqtt_demo.axf._ld\"\n")
	loggers.Log("\n")
	loggers.LogError("\n")
	err = loggers.SetLogSource("source2")
	require.Nil(t, err)

	loggers.Log("Test output3")
	loggers.LogError("Test err1")
	err = loggers.SetLogSource("source3")
	require.Nil(t, err)

	err = loggers.SetLoggerSource("LoggerSource2")
	require.Nil(t, err)

	loggers.LogError("Test err2")
	err = loggers.SetLogSource("source4")
	require.Nil(t, err)

	loggers.LogError("Test err3")
	err = loggers.Close()
	require.Nil(t, err)
}
