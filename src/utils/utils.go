package utils

import (
	"errors"
	"slices"
	"time"
	"timetrack-sync/src/sloneek"
	toggltrack "timetrack-sync/src/togglTrack"

	"github.com/rs/zerolog"
)

// TODO tohle by asi davalo smysl mit jako metodu na business entite, aby se nemusel kontrolovat ten pointer
func RoundTimeEntry(entry *toggltrack.TimeEntry) (*toggltrack.TimeEntry, error) {
	if entry == nil {
		return nil, errors.New("Time entry may not be nil")
	}

	entry.Start = entry.Start.Round(15 * time.Minute)
	entry.Stop = entry.Stop.Round(15 * time.Minute)

	return entry, nil
}

func ParseDateString(value string, logger *zerolog.Logger, errorMessage *string) time.Time {
	message := "Error while parsing date"
	if errorMessage != nil {
		message = *errorMessage
	}

	timeValue, err := time.Parse(time.DateOnly, value)
	if err != nil {
		logger.Fatal().Err(err).Msg(message)
	}

	return timeValue
}

func ParseDateTimeString(value string, logger *zerolog.Logger, errorMessage *string) time.Time {
	message := "Error while parsing date time"
	if errorMessage != nil {
		message = *errorMessage
	}

	timeValue, err := time.Parse(time.DateTime, value)
	if err != nil {
		logger.Fatal().Err(err).Msg(message)
	}

	return timeValue
}

func MapTogglProjectToSloneekActivityAndCategory(project string) (string, string) {
	// TODO this as an external config?
	switch project {
	case "Proteus":
		return "Vývoj", "Proteus"
	case "Copilot":
		return "Vývoj", "Proteus"
	case "Portál":
		return "Vývoj", "Portál"
	case "Akvizice":
		return "Vývoj", "Akviziční formulář"
	case "Flexi":
		return "Vývoj", "Flexi"
	case "Interní":
		return "Vývoj", "Iternal job"
	case "Hiring":
		return "Hiring", ""
	case "Admin & Meetings":
		return "Meeting", ""
	}

	return "", ""
}

func MapTogglEntryToSloneekEntry(
	entry *toggltrack.TimeEntry,
	togglProjects []toggltrack.Project,
	sloneekActivities []sloneek.Activity,
	sloneekCategories []sloneek.Category,
	logger *zerolog.Logger,
) (*sloneek.TimeEntry, error) {
	logger.Debug().Any("entry", entry).Msg("Mapping toggl entry to sloneek entry")
	projectIndex := slices.IndexFunc(togglProjects, func(project toggltrack.Project) bool { return project.Id == *entry.ProjectID })
	if projectIndex == -1 {
		logger.Error().Any("entry", *entry).Msg("Project for entry not found")
		return nil, errors.New("Project for entry not found")
	}

	project := togglProjects[projectIndex]
	activityName, categoryName := MapTogglProjectToSloneekActivityAndCategory(project.Name)
	if activityName == "" {
		logger.Error().Str("project", project.Name).Msg("Could not find matching activity")
		return nil, errors.New("Could not find matching activity")
	}

	activityIndex := slices.IndexFunc(sloneekActivities, func(activity sloneek.Activity) bool { return activity.Name == activityName })
	if activityIndex == -1 {
		logger.Error().Str("activity", activityName).Msg("Activity not found")
		return nil, errors.New("Activity not found")
	}

	activity := &sloneekActivities[activityIndex]

	var category *sloneek.Category
	if categoryName != "" {
		categoryIndex := slices.IndexFunc(sloneekCategories, func(category sloneek.Category) bool { return category.Name == categoryName })
		if categoryIndex == -1 {
			logger.Error().Str("category", activityName).Msg("Category not found")
			// teoreticky by stacilo pokracovat jen s aktivitou, ale pro ted radeji konec
			return nil, errors.New("Category not found")
		}

		category = &sloneekCategories[categoryIndex]
	}

	var categoryId *string
	if category != nil {
		categoryId = &category.Id
	}

	sloneekEntry := &sloneek.TimeEntry{
		ActivityId: activity.Id,
		CategoryId: categoryId,
		Since:      entry.Start,
		Until:      entry.Stop,
	}

	return sloneekEntry, nil
}
