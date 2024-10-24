package sloneek

import (
	"testing"
	testutils "timetrack-sync/src/testUtils"
)

func TestGetHoursReturnsCorrectFullHour(t *testing.T) {
	expectedHours := float64(1)
	entry := &TimeEntry{
		ActivityId: "1",
		Since:      testutils.DateTimeFromString("2024-01-01 00:00:00", t),
		Until:      testutils.DateTimeFromString("2024-01-01 01:00:00", t),
	}

	result := entry.GetHours()

	if result != expectedHours {
		t.Errorf("Unexpected hours value found. Expected: %f, got: %f", result, expectedHours)
	}
}

func TestGetHoursReturnsCorrectPartialHour(t *testing.T) {
	expectedHours := float64(1.25)
	entry := &TimeEntry{
		ActivityId: "1",
		Since:      testutils.DateTimeFromString("2024-01-01 00:00:00", t),
		Until:      testutils.DateTimeFromString("2024-01-01 01:15:00", t),
	}

	result := entry.GetHours()

	if result != expectedHours {
		t.Errorf("Unexpected hours value found. Expected: %f, got: %f", result, expectedHours)
	}
}
