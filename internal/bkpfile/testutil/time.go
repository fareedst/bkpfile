package testutil

import (
	"testing"
	"time"
)

// MockTimeNow replaces the time.Now function with a fixed time for testing
func MockTimeNow(t *testing.T, fixedTime time.Time) func() {
	originalTimeNow := timeNow
	timeNow = func() time.Time { return fixedTime }
	return func() { timeNow = originalTimeNow }
}

// timeNow is a variable that can be replaced in tests to provide a fixed time
var timeNow = time.Now
