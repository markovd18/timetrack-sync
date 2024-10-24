package testutils

import (
	"testing"
	"time"
)

func DateTimeFromString(value string, t *testing.T) time.Time {
	t.Helper()
	result, err := time.Parse(time.DateTime, value)
	if err != nil {
		t.Errorf("PArsing failed: %v", err)
	}

	return result
}
