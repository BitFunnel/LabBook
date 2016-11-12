package labtest

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/BitFunnel/LabBook/src/systems"
	"github.com/stretchr/testify/assert"
)

// AssertEventsEqual checks that two event logs are equal.
func AssertEventsEqual(t *testing.T, expectedEventLog []string, actualEventLog []systems.Operation) bool {
	if len(expectedEventLog) != len(actualEventLog) {
		return logEventLogsAndFail(t, expectedEventLog, actualEventLog)
	}

	for i := range expectedEventLog {
		if expectedEventLog[i] != actualEventLog[i].String() {
			return logEventLogsAndFail(t, expectedEventLog, actualEventLog)
		}
	}

	return true
}

// LogEventLogsAndFail will log event logs and fail.
func logEventLogsAndFail(t *testing.T, expectedEventLog []string, actualEventLog []systems.Operation) bool {
	expectedString := strings.Join(expectedEventLog, "\n")
	actualString := eventLogString(t, actualEventLog)
	return assert.Fail(
		t,
		fmt.Sprintf(
			"EXPECTED:\n%s\n\nACTUAL:\n%s\n",
			expectedString,
			actualString))
}

// EventLogString outputs an event log to the test harness's logging facilities.
func eventLogString(t *testing.T, eventLog []systems.Operation) string {
	var buffer bytes.Buffer
	for i := range eventLog {
		buffer.WriteString(fmt.Sprintf("%s\n", eventLog[i].String()))
	}
	return buffer.String()
}
