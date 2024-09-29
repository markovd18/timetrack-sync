package utils

import (
	"fmt"
	"testing"
	"time"
	"timetrack-sync/src/sloneek"
	toggltrack "timetrack-sync/src/togglTrack"

	"github.com/rs/zerolog"
)

func dateTimeFromString(value string, t *testing.T) time.Time {
	t.Helper()
	result, err := time.Parse(time.DateTime, value)
	if err != nil {
		t.Errorf("PArsing failed: %v", err)
	}

	return result
}

func TestRoundTimeEntryWorksAsExpected(t *testing.T) {

	testCases := []struct {
		Start         string
		End           string
		ExpectedStart string
		ExpectedEnd   string
	}{
		{Start: "2024-01-01 10:01:00", End: "2024-01-01 10:12:00", ExpectedStart: "2024-01-01 10:00:00", ExpectedEnd: "2024-01-01 10:15:00"},
		{Start: "2024-01-01 10:00:00", End: "2024-01-01 10:15:00", ExpectedStart: "2024-01-01 10:00:00", ExpectedEnd: "2024-01-01 10:15:00"},
		{Start: "2024-01-01 09:59:58", End: "2024-01-01 10:08:00", ExpectedStart: "2024-01-01 10:00:00", ExpectedEnd: "2024-01-01 10:15:00"},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("Start %s, End %s", testCase.Start, testCase.End), func(t *testing.T) {
			start := dateTimeFromString(testCase.Start, t)
			end := dateTimeFromString(testCase.End, t)

			expectedStart := dateTimeFromString(testCase.ExpectedStart, t)
			expectedEnd := dateTimeFromString(testCase.ExpectedEnd, t)

			entry := toggltrack.TimeEntry{ID: 1, Start: start, Stop: end, TaskID: 10, Duration: 15, Description: "test"}
			result, testErr := RoundTimeEntry(&entry)
			if testErr != nil {
				t.Error(testErr)
			}

			if !result.Start.Equal(expectedStart) {
				t.Errorf("Expected %v. got %v.", expectedStart, result.Start)
			}

			if !result.Stop.Equal(expectedEnd) {
				t.Errorf("Expected %v. got %v.", expectedEnd, result.Stop)
			}
		})

	}

}

func TestMapToggleToSloneekEntryFailsWhenProjectNotFound(t *testing.T) {

	activities := []sloneek.Activity{
		{Id: "1", Name: "Vývoj"},
		{Id: "2", Name: "Hiring"},
		{Id: "3", Name: "Meeting"},
	}

	categories := []sloneek.Category{
		{Id: "1", Name: "Proteus"},
		{Id: "2", Name: "Portál"},
		{Id: "3", Name: "Akviziční formulář"},
		{Id: "4", Name: "Flexi"},
	}

	projects := []toggltrack.Project{
		{Name: "Proteus", Id: 1},
		{Name: "Akvizice", Id: 2},
		{Name: "Portál", Id: 3},
		{Name: "Hiring", Id: 4},
		{Name: "Flexi", Id: 5},
		{Name: "Copilot", Id: 6},
	}

	projectId := int32(10)
	togglEntry := toggltrack.TimeEntry{
		ID:        1,
		Start:     dateTimeFromString("2024-01-01 10:00:00", t),
		Stop:      dateTimeFromString("2024-01-01 10:15:00", t),
		ProjectID: &projectId,
		Duration:  15 * 60,
	}

	_, err := MapTogglEntryToSloneekEntry(&togglEntry, projects, activities, categories, &zerolog.Logger{})
	if err == nil {
		t.Errorf("Expected to fail but did not fail")
	}
}

func TestMapToggleToSloneekEntryFailsWhenActivityNotFound(t *testing.T) {

	activities := []sloneek.Activity{
		{Id: "1", Name: "Vývoj"},
		{Id: "2", Name: "Hiring"},
		{Id: "3", Name: "Meeting"},
	}

	categories := []sloneek.Category{
		{Id: "1", Name: "Proteus"},
		{Id: "2", Name: "Portál"},
		{Id: "3", Name: "Akviziční formulář"},
		{Id: "4", Name: "Flexi"},
	}

	projects := []toggltrack.Project{
		{Name: "Protezus", Id: 1},
		{Name: "Akvizice", Id: 2},
		{Name: "Portál", Id: 3},
		{Name: "Hiring", Id: 4},
		{Name: "Flexi", Id: 5},
		{Name: "Copilot", Id: 6},
	}

	projectId := int32(1)
	togglEntry := toggltrack.TimeEntry{
		ID:        1,
		Start:     dateTimeFromString("2024-01-01 10:00:00", t),
		Stop:      dateTimeFromString("2024-01-01 10:15:00", t),
		ProjectID: &projectId,
		Duration:  15 * 60,
	}

	_, err := MapTogglEntryToSloneekEntry(&togglEntry, projects, activities, categories, &zerolog.Logger{})
	if err == nil {
		t.Errorf("Expected to fail but did not fail")
	}
}

func TestMapToggleToSloneekEntryFailsWHenCategoryNotFound(t *testing.T) {

	activities := []sloneek.Activity{
		{Id: "1", Name: "Vývoj"},
		{Id: "2", Name: "Hiring"},
		{Id: "3", Name: "Meeting"},
	}

	categories := []sloneek.Category{
		{Id: "1", Name: "Protezus"},
		{Id: "2", Name: "Portál"},
		{Id: "3", Name: "Akviziční formulář"},
		{Id: "4", Name: "Flexi"},
	}

	projects := []toggltrack.Project{
		{Name: "Proteus", Id: 1},
		{Name: "Akvizice", Id: 2},
		{Name: "Portál", Id: 3},
		{Name: "Hiring", Id: 4},
		{Name: "Flexi", Id: 5},
		{Name: "Copilot", Id: 6},
	}

	projectId := int32(1)
	togglEntry := toggltrack.TimeEntry{
		ID:        1,
		Start:     dateTimeFromString("2024-01-01 10:00:00", t),
		Stop:      dateTimeFromString("2024-01-01 10:15:00", t),
		ProjectID: &projectId,
		Duration:  15 * 60,
	}

	_, err := MapTogglEntryToSloneekEntry(&togglEntry, projects, activities, categories, &zerolog.Logger{})
	if err == nil {
		t.Errorf("Expected to fail but did not fail")
	}
}

func TestMapToggleToSloneekWorksAsExpected(t *testing.T) {
	expectedActivityId := "1"
	expectedCategoryId := "1"
	activities := []sloneek.Activity{
		{Id: expectedActivityId, Name: "Vývoj"},
		{Id: "2", Name: "Hiring"},
		{Id: "3", Name: "Meeting"},
	}

	categories := []sloneek.Category{
		{Id: expectedCategoryId, Name: "Proteus"},
		{Id: "2", Name: "Portál"},
		{Id: "3", Name: "Akviziční formulář"},
		{Id: "4", Name: "Flexi"},
	}

	projectId := int32(1)
	projects := []toggltrack.Project{
		{Name: "Proteus", Id: projectId},
		{Name: "Akvizice", Id: 2},
		{Name: "Portál", Id: 3},
		{Name: "Hiring", Id: 4},
		{Name: "Flexi", Id: 5},
		{Name: "Copilot", Id: 6},
	}

	entryStart := dateTimeFromString("2024-01-01 10:00:00", t)
	entryStop := dateTimeFromString("2024-01-01 10:15:00", t)
	togglEntry := toggltrack.TimeEntry{
		ID:        1,
		Start:     entryStart,
		Stop:      entryStop,
		ProjectID: &projectId,
		Duration:  15 * 60,
	}

	result, err := MapTogglEntryToSloneekEntry(&togglEntry, projects, activities, categories, &zerolog.Logger{})
	if err != nil {
		t.Errorf("Mapping function returned unexpected error: %v", err)
	}

	if result.ActivityId != expectedActivityId {
		t.Errorf("Unexpected activity found. Expected %s, got %s", expectedActivityId, result.ActivityId)
	}
	if *result.CategoryId != expectedCategoryId {
		t.Errorf("Unexpected category found. Expected %s, got %s", expectedCategoryId, *result.CategoryId)
	}
	if !(*result).Since.Equal(entryStart) {
		t.Errorf("Unexpected start time found. Expected %s, got %s", entryStart, (*result).Since)
	}
	if !(*result).Until.Equal(entryStop) {
		t.Errorf("Unexpected end time found. Expected %s, got %s", entryStart, (*result).Since)
	}
}

func TestMapToggleToSloneekWorksAsExpectedSecond(t *testing.T) {
	expectedActivityId := "2"
	activities := []sloneek.Activity{
		{Id: expectedActivityId, Name: "Vývoj"},
		{Id: "2", Name: "Hiring"},
		{Id: "3", Name: "Meeting"},
	}

	categories := []sloneek.Category{
		{Id: "1", Name: "Proteus"},
		{Id: "2", Name: "Portál"},
		{Id: "3", Name: "Akviziční formulář"},
		{Id: "4", Name: "Flexi"},
	}

	projectId := int32(4)
	projects := []toggltrack.Project{
		{Name: "Proteus", Id: projectId},
		{Name: "Akvizice", Id: 2},
		{Name: "Portál", Id: 3},
		{Name: "Hiring", Id: projectId},
		{Name: "Flexi", Id: 5},
		{Name: "Copilot", Id: 6},
	}

	entryStart := dateTimeFromString("2024-01-01 10:00:00", t)
	entryStop := dateTimeFromString("2024-01-01 10:15:00", t)
	togglEntry := toggltrack.TimeEntry{
		ID:        1,
		Start:     entryStart,
		Stop:      entryStop,
		ProjectID: &projectId,
		Duration:  15 * 60,
	}

	result, err := MapTogglEntryToSloneekEntry(&togglEntry, projects, activities, categories, &zerolog.Logger{})
	if err != nil {
		t.Errorf("Mapping function returned unexpected error: %v", err)
	}

	if result.ActivityId != expectedActivityId {
		t.Errorf("Unexpected activity found. Expected %s, got %s", expectedActivityId, result.ActivityId)
	}
	if result.CategoryId != nil {
		t.Errorf("Unexpected category found. Expected %v, got %s", nil, *result.CategoryId)
	}
	if !(*result).Since.Equal(entryStart) {
		t.Errorf("Unexpected start time found. Expected %s, got %s", entryStart, (*result).Since)
	}
	if !(*result).Until.Equal(entryStop) {
		t.Errorf("Unexpected end time found. Expected %s, got %s", entryStart, (*result).Since)
	}
}
