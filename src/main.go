package main

import (
	"flag"
	"slices"
	"timetrack-sync/src/sloneek"
	toggltrack "timetrack-sync/src/togglTrack"
	"timetrack-sync/src/utils"

	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

var (
	SLONEEK_API   = "https://api2.sloneek.com"
	TOGGL_API_URL = "https://api.track.toggl.com/api/v9"
)

func RoundTimeEntries(entries []toggltrack.TimeEntry) []toggltrack.TimeEntry {
	entriesLen := len(entries)
	for i := 0; i < entriesLen; i++ {
		utils.RoundTimeEntry(&entries[i])
	}

	return entries
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	output := os.Stderr
	logger := zerolog.New(output).With().Timestamp().Logger().Level(zerolog.InfoLevel).Output(zerolog.ConsoleWriter{Out: output, TimeFormat: time.StampMilli})

	logger.Info().Msg("Loading environment variables")
	err := godotenv.Load(".env")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error while loading environment variables")
	}

	bearerToken := flag.String("bearer", "", "Bearer token obtained after login to Sloneek app")
	dryRun := flag.Bool("dry-run", false, "Whether or not to launch a dry run which does not persist any data.")

	logger.Info().Msg("Parsing CLI flags")
	flag.Parse()

	if *bearerToken == "" {
		logger.Fatal().Msg("Sloneek JWT not found")
		flag.Usage()
		return
	}

	// TODO CLI flag
	togglApiKey := os.Getenv("TOGGL_API_KEY")
	togglLogger := logger.With().Str("client", "toggl").Logger()
	togglTrackClient := toggltrack.CreateTogglTrackClient(TOGGL_API_URL, togglApiKey, &togglLogger)

	// TODO CLI flagy
	errorMsg := "Error while parsing interval start."
	since := utils.ParseDateString("2024-09-01", &logger, &errorMsg)

	errorMsg = "Error while parsing interval end."
	until := utils.ParseDateString("2024-10-01", &logger, &errorMsg)

	togglTimeEntries := togglTrackClient.GetTimeEntries(since, until)

	logger.Info().Msg("Rounding time entries")
	roundedEntries := RoundTimeEntries(togglTimeEntries)
	togglProjects := togglTrackClient.GetProjects()

	sloneekLogger := logger.With().Str("client", "sloneek").Logger()
	sloneekClient := sloneek.CreateSloneekClient(SLONEEK_API, *bearerToken, &sloneekLogger)

	sloneekCategories := sloneekClient.GetCategories()
	sloneekActivities := sloneekClient.GetActivities()

	logger.Info().Msg("Mapping Toggl time entries to Sloneek time entries")
	sloneekEntries := []sloneek.TimeEntry{}
	for _, entry := range roundedEntries {
		sloneekEntry, err := utils.MapTogglEntryToSloneekEntry(&entry, togglProjects, sloneekActivities, sloneekCategories, &logger)
		if err != nil {
			logger.Fatal().Err(err).Msg("Error while mapping toggle entry to sloneek entry")
		}

		logger.Debug().Any("sloneek_entry", sloneekEntry).Any("toggl_entry", entry).Msg("Entry mapped.")
		sloneekEntries = append(sloneekEntries, *sloneekEntry)
	}

	logger.Debug().Any("result", sloneekEntries).Msg("Mam vysledek")

	if dryRun != nil && !*dryRun {
		logger.Info().Msg("Sending time entries to Sloneek")
		for _, entry := range sloneekEntries {
			err := sloneekClient.SaveTimeEntry(&entry)
			if err != nil {
				logger.Error().Err(err).Any("entry", entry).Msg("Failed to save Sloneek time entry")
				break
			}
		}
	}

	activityTotalTimesMap := make(map[string]float64)
	for _, entry := range sloneekEntries {
		projectId := entry.GetProjectId()
		activityTotalHours := activityTotalTimesMap[projectId]

		activityTotalTimesMap[projectId] = activityTotalHours + entry.GetHours()
	}

	logger.Info().Msg("Total Sloneek hours summary:")
	totalHours := float64(0)
	for projectId, activityHours := range activityTotalTimesMap {
		projectName := ""
		categoryIndex := slices.IndexFunc(sloneekCategories, func(category sloneek.Category) bool { return category.Id == projectId })
		if categoryIndex != -1 {
			category := &sloneekCategories[categoryIndex]
			projectName = category.Name
		}

		// category not found
		if projectName == "" {
			activityIndex := slices.IndexFunc(sloneekActivities, func(activity sloneek.Activity) bool { return activity.Id == projectId })
			if activityIndex == -1 {
				logger.Fatal().Str("activityId", projectId).Msg("Activity not found")
			}

			activity := &sloneekActivities[activityIndex]
			projectName = activity.Name
		}

		logger.Info().Msgf("%s: %f hours.", projectName, activityHours)
		totalHours += activityHours
	}

	logger.Info().Msgf("Total hours: %f", totalHours)

	// TODO project overview - summary, times etc.
}
