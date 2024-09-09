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

func MapTogglProjectToSloneekActivityAndCategory(project string) (string, string) {
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

	//{
	//"isRepeat":false,
	//"user_planning_event_uuid":"53016fa7-44fd-4726-949e-b8b66b72c38c",
	//"planning_categories":["6adf0954-5cf0-4de7-b161-30a4f4d4caec"],
	//"started_at":"2024-08-05T09:00:00+02:00",
	//"ended_at":"2024-08-05T09:30:00+02:00",
	//"start_time":"09:00:00+02:00",
	//"end_time":"09:30:00+02:00",
	//"days":[],
	//"duration_time":"2024-08-04T22:30:00.000Z",
	//"duration":30,
	//"timezone":"2024-08-31T17:20:56+02:00",
	//"note":"",
	//"is_automatically_approve":false,
	//"message":"",
	//"mentions":[],
	//"client":"",
	//"client_project":""
	//}
}
