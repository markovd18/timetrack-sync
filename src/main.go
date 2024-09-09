package main

import (
	"encoding/json"
	"errors"
	"flag"
	toggltrack "timetrack-sync/src/togglTrack"
	"timetrack-sync/src/utils"

	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

var (
	SLONEEK_API   = "https://api2.sloneek.com"
	TOGGL_API_URL = "https://api.track.toggl.com/api/v9"
)

type Planning_Event struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
	// pak este neco
}

type User_Planning_event struct {
	Uuid           string         `json:"uuid"`
	Planning_Event Planning_Event `json:"planning_event"`
}

type Options_Response struct {
	Message     string                `json:"message"`
	Status_code int32                 `json:"status_code"`
	Data        []User_Planning_event `json:"data"`
}

type Category struct {
	Uuid string `json:"uuid"`
	Name string `jons:"name"`
}

type Categories_Response struct {
	Message     string     `json:"message"`
	Status_code int32      `json:"status_code"`
	Data        []Category `json:"data"`
}

// TimeEntry represents a time entry in Toggl Track.
type TimeEntry struct {
	ID          int64     `json:"id"`
	ProjectID   *int64    `json:"project_id,omitempty"`
	TaskID      int64     `json:"task_id"`
	Start       time.Time `json:"start"`
	Stop        time.Time `json:"stop,omitempty"`
	Duration    int64     `json:"duration"`
	Description string    `json:"description"`
	//Tags        []string   `json:"tags"`
	//TagIDs      []int64    `json:"tag_ids"`
}

func get_request(url string, logger *zerolog.Logger) []byte {
	res, err := http.Get(url)
	if err != nil {
		logger.Fatal().Err(err).Msg("Doslo k nejaky chybicce")
	}

	fmt.Printf("response status: %v\n", res.Status)
	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se precist body")
	}

	return body
}

func test_api(logger *zerolog.Logger) {
	body := get_request(SLONEEK_API, logger)
	fmt.Printf("body: %s\n", body)
}

func vic_test_sloneek(logger *zerolog.Logger, bearer_token *string) {
	user_uuid := os.Getenv("USER_UUID")
	endpoint_url := fmt.Sprintf("%s/v2/module-planning/scheduled-events/options/user-planning-events?user_uuid=%s", SLONEEK_API, user_uuid)
	req, err := http.NewRequest(http.MethodGet, endpoint_url, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se udelat request")
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *bearer_token))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se poslat auth request")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se precist body")
	}

	logger.Info().Str("body", fmt.Sprintf("%s", body)).Msg("Prisla mi odpoved")
	var payload Options_Response
	err = json.Unmarshal(body, &payload)
	if err != nil {
		logger.Fatal().Err(err).Msg("Unmarshal se nepovedl")
	}

	fmt.Printf("payload.Data: %v\n", payload.Data)

	activities_map := make(map[string]string)
	for _, item := range payload.Data {
		activities_map[item.Planning_Event.Name] = item.Planning_Event.Uuid
	}

	fmt.Printf("activities_map: %v\n", activities_map)

	categories_url := fmt.Sprintf("%s/v2/module-planning/scheduled-events/options/categories", SLONEEK_API)
	req_categories, err := http.NewRequest(http.MethodGet, categories_url, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se udelat categories request")
	}

	req_categories.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *bearer_token))
	res_categories, err := http.DefaultClient.Do(req_categories)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se poslat categories request")
	}

	body, err = io.ReadAll(res_categories.Body)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se precist categories body")
	}

	logger.Info().Str("body", fmt.Sprintf("%s", body)).Msg("Prisla mi categories odpoved")
	var categories_payload Categories_Response
	err = json.Unmarshal(body, &categories_payload)
	if err != nil {
		logger.Fatal().Err(err).Msg("Unmarshal se nepovedl")
	}

	fmt.Printf("payload.Data: %v\n", categories_payload.Data)
	categories_map := make(map[string]string)
	for _, item := range categories_payload.Data {
		categories_map[item.Name] = item.Uuid
	}
	fmt.Printf("categories map: %v\n", categories_map)
}

func read_body(body_stream io.ReadCloser, logger *zerolog.Logger) []byte {
	body, err := io.ReadAll(body_stream)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se precist categories body")
	}

	return body
}

func RoundTimeEntries(entries []toggltrack.TimeEntry) []toggltrack.TimeEntry {
	for _, timeEntry := range entries {
		utils.RoundTimeEntry(&timeEntry)
	}

	return entries
}

// TODO tohle by asi davalo smysl mit jako metodu na business entite, aby se nemusel kontrolovat ten pointer
func RoundTimeEntry(entry *TimeEntry) (*TimeEntry, error) {
	if entry == nil {
		return nil, errors.New("Time entry may not be nil")
	}

	entry.Start = entry.Start.Round(15 * time.Minute)
	entry.Stop = entry.Stop.Round(15 * time.Minute)

	return entry, nil
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	output := os.Stderr
	logger := zerolog.New(output).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: output, TimeFormat: time.StampMilli})

	err := godotenv.Load(".env")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error while loading environment variables")
	}

	bearer_token := flag.String("bearer", "", "Bearer token obtained after login to Sloneek app")

	flag.Parse()

	if *bearer_token == "" {
		logger.Fatal().Msg("Nezadal jsi JWT")
		return
	}

	//sloneekClient := sloneek.CreateSloneekClient(SLONEEK_API, *bearer_token, &logger)
	//sloneekCategories := sloneekClient.GetCategories()
	//sloneedActivities := sloneekClient.GetActivities()

	togglApiKey := os.Getenv("TOGGL_API_KEY")
	togglTrackClient := toggltrack.CreateTogglTrackClient(TOGGL_API_URL, togglApiKey, &logger)
	since, err := time.Parse(time.DateOnly, "2024-08-01")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error pri parsovani zacatku intervalu")
	}

	until, err := time.Parse(time.DateOnly, "2024-08-02")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error pri parsovani konce intervalu")
	}

	togglTimeEntries := togglTrackClient.GetTimeEntries(since, until)
	fmt.Printf("togglTimeEntries: %v\n", togglTimeEntries)

	//togglProjects := togglTrackClient.GetProjects()

	RoundTimeEntries(togglTimeEntries)

}
