package toggltrack

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type TimeEntry struct {
	ID          int64     `json:"id"`
	ProjectID   *int32    `json:"project_id,omitempty"`
	TaskID      int64     `json:"task_id"`
	Start       time.Time `json:"start"`
	Stop        time.Time `json:"stop,omitempty"`
	Duration    int64     `json:"duration"`
	Description string    `json:"description"`
}

type TogglTrackClient struct {
	apiUrl     string
	apiKey     string
	logger     *zerolog.Logger
	httpClient *http.Client
}

func CreateTogglTrackClient(apiUrl string, apiKey string, logger *zerolog.Logger) *TogglTrackClient {
	logger.Info().Msg("Initializing Toggl client")
	httpClient := http.Client{Timeout: time.Minute}
	return &TogglTrackClient{apiUrl: apiUrl, apiKey: apiKey, logger: logger, httpClient: &httpClient}
}

func (client *TogglTrackClient) authenticateRequest(req *http.Request) {
	req.SetBasicAuth(client.apiKey, "api_token")
}

func (client *TogglTrackClient) GetTimeEntries(since time.Time, until time.Time) []TimeEntry {
	client.logger.Info().Any("since", since).Any("until", until).Msg("Looking up Toggl time entries")
	time_entries_url := fmt.Sprintf("%s/me/time_entries?start_date=%s&end_date=%s", client.apiUrl, since.Format(time.DateOnly), until.Format(time.DateOnly))
	req, err := http.NewRequest(http.MethodGet, time_entries_url, nil)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while creating request.")
	}

	client.authenticateRequest(req)

	res, err := client.httpClient.Do(req)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while sending request.")
	}

	client.logger.Debug().Str("status", res.Status).Msg("Prisla mi repsonse")
	if res.StatusCode == 403 || res.StatusCode == 401 {
		client.logger.Fatal().Str("api_key", client.apiKey).Msg("Authentication failed")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while reading response body.")
	}
	client.logger.Debug().Str("response_body", fmt.Sprintf("%s", body)).Msg("Prislo mi response body")

	var time_entries []TimeEntry
	err = json.Unmarshal(body, &time_entries)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while unmarshaling response payload.")
	}

	client.logger.Info().Msg("Returning time entries.")
	return time_entries
}

type MePayload struct {
	DefaultWorkspaceId int32 `json:"default_workspace_id"`
}

func (client *TogglTrackClient) GetDefaultWorkspaceId() int32 {
	client.logger.Info().Msg("Looking up dwfault Toggl workspace ID")
	meUrl := fmt.Sprintf("%s/me", client.apiUrl)
	req, err := http.NewRequest(http.MethodGet, meUrl, nil)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se vytvorit request")
	}

	client.authenticateRequest(req)

	res, err := client.httpClient.Do(req)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se poslat auth request")
	}

	client.logger.Info().Str("status", res.Status).Msg("Prisla mi repsonse")
	if res.StatusCode == 403 || res.StatusCode == 401 {
		client.logger.Fatal().Str("api_key", client.apiKey).Msg("Autentizace selhala")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se precist categories body")
	}

	client.logger.Info().Str("response_body", fmt.Sprintf("%s", body)).Msg("Prislo mi response body")
	var payload MePayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error pri unmarshalingu toggl responsu")
	}

	return payload.DefaultWorkspaceId

}

type Project struct {
	Name string `json:"name"`
	Id   int32  `json:"id"`
}

func (client *TogglTrackClient) GetProjects() []Project {
	client.logger.Info().Msg("Looking up Toggl projects")
	defaultWorkpaceId := client.GetDefaultWorkspaceId()
	projectsUrl := fmt.Sprintf("%s/workspaces/%d/projects", client.apiUrl, defaultWorkpaceId)
	req, err := http.NewRequest(http.MethodGet, projectsUrl, nil)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while creating request.")
	}

	client.authenticateRequest(req)

	res, err := client.httpClient.Do(req)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while sending request.")
	}

	client.logger.Debug().Str("status", res.Status).Msg("Prisla mi repsonse")
	if res.StatusCode == 403 || res.StatusCode == 401 {
		client.logger.Fatal().Msg("Authentication failed")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while reading response body.")
	}

	client.logger.Debug().Str("response_body", fmt.Sprintf("%s", body)).Msg("Prislo mi response body")
	var projects []Project
	err = json.Unmarshal(body, &projects)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while unmarshaling resonse payload.")
	}

	return projects
}
