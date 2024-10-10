package main

import (
	"flag"
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
	since := utils.ParseDateString("2024-09-30", &logger, &errorMsg)

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

	// test sloneek postu
	//categoryId := "8969d1b2-57bb-4bb4-8174-8b39b8dd6dcb"
	//sloneekClient.SaveTimeEntry(&sloneek.TimeEntry{
	//	ActivityId: "d04f6cb7-1186-4d04-9d84-c23501fe4ae9",
	//	CategoryId: &categoryId,
	//	Since:      utils.ParseDateTimeString("2024-09-29 10:00:00", &logger, nil),
	//	Until:      utils.ParseDateTimeString("2024-09-29 10:30:00", &logger, nil),
	//})

	logger.Info().Msg("Sending time entries to Sloneek")

	for _, entry := range sloneekEntries {
		err := sloneekClient.SaveTimeEntry(&entry)
		if err != nil {
			logger.Error().Err(err).Any("entry", entry).Msg("Failed to save Sloneek time entry")
			break
		}
	}

	// TODO project overview - summary, times etc.
}
