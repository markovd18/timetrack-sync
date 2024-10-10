package sloneek

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type SloneekClient struct {
	apiUrl      string
	bearerToken string
	httpClient  *http.Client
	logger      *zerolog.Logger
}

type Category struct {
	Id   string `json:"uuid"`
	Name string `jons:"name"`
}

type CategoriesResponse struct {
	Message     string     `json:"message"`
	Status_code int32      `json:"status_code"`
	Data        []Category `json:"data"`
}

// zajimaj me hlavne "Meeting", "Hiring", "Vývoj"
func (client *SloneekClient) GetCategories() []Category {
	client.logger.Info().Msg("Looking up Sloneek categories")
	categoriesUrl := fmt.Sprintf("%s/v2/module-planning/scheduled-events/options/categories", client.apiUrl)
	req, err := http.NewRequest(http.MethodGet, categoriesUrl, nil)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se udelat categories request")
		return nil
	}

	client.authenticateRequest(req)
	res, err := client.httpClient.Do(req)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se poslat categories request")
		return nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while reading response body.")
		return nil
	}

	client.logger.Debug().Str("body", fmt.Sprintf("%s", body)).Msg("Prisla mi categories odpoved")
	var categoriesPayload CategoriesResponse
	err = json.Unmarshal(body, &categoriesPayload)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while unmarshaling response body payload.")
		return nil
	}

	categories := categoriesPayload.Data
	client.logger.Debug().Any("sloneek_categories", categories).Msg("Got sloneek categories")
	client.logger.Info().Msg("Categories found.")
	return categories
}

type PlanningEvent struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
}

type UserPlanningEvent struct {
	Uuid           string        `json:"uuid"`
	Planning_Event PlanningEvent `json:"planning_event"`
}

type OptionsResponse struct {
	Message     string              `json:"message"`
	Status_code int32               `json:"status_code"`
	Data        []UserPlanningEvent `json:"data"`
}

type Activity struct {
	Id   string
	Name string
}

func (client *SloneekClient) GetActivities() []Activity {
	client.logger.Info().Msg("Looking up Sloneek activities")
	endpointUrl := fmt.Sprintf("%s/v2/module-planning/scheduled-events/options/user-planning-events", client.apiUrl)
	req, err := http.NewRequest(http.MethodGet, endpointUrl, nil)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while creating request.")
		return nil
	}

	client.authenticateRequest(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while sending request.")
		return nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while reading response body.")
		return nil
	}

	client.logger.Debug().Str("body", fmt.Sprintf("%s", body)).Msg("Prisla mi odpoved")
	var payload OptionsResponse
	err = json.Unmarshal(body, &payload)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while unmarshaling response payload.")
	}

	client.logger.Debug().Any("payload_data", payload.Data)

	activities := make([]Activity, len(payload.Data))
	for i, item := range payload.Data {
		// we need to set the User Planning event UUID, not the underlying planning event ID
		// since planning event UUID is shared, user planning event is specific to the user
		activities[i] = Activity{Id: item.Uuid, Name: item.Planning_Event.Name}
	}

	client.logger.Debug().Any("sloneek_activities", activities).Msg("Got sloneek activities")
	client.logger.Info().Msg("Activities found.")
	return activities

}

func CreateSloneekClient(apiUrl string, bearerToken string, logger *zerolog.Logger) *SloneekClient {
	logger.Info().Msg("Initializing Sloneek client")
	httpClient := http.Client{Timeout: time.Minute}
	return &SloneekClient{apiUrl: apiUrl, bearerToken: bearerToken, logger: logger, httpClient: &httpClient}
}

func (client *SloneekClient) authenticateRequest(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.bearerToken))
}

type TimeEntry struct {
	ActivityId string
	CategoryId *string
	note       string
	Since      time.Time
	Until      time.Time
}

type TimeEntryDTO struct {
	UserPlanningEventUuid  string    `json:"user_planning_event_uuid"`
	PlanningCategories     []string  `json:"planning_categories"`
	StartedAt              time.Time `json:"started_at"`
	EndedAt                time.Time `json:"ended_at"`
	StartTime              time.Time `json:"start_time"`
	EndTime                time.Time `json:"end_time"`
	Note                   string    `json:"note"`
	IsAutomaticallyApprove bool      `json:"is_automatically_approve"`
}

func (client *SloneekClient) SaveTimeEntry(timeEntry *TimeEntry) error {
	client.logger.Info().Any("time_entry", timeEntry).Msg("Saving Sloneek time entry")
	if timeEntry == nil {
		err := errors.New("Time entry to save may not be nil")
		client.logger.Err(err).Msg("Error while saving time entry")
		return err
	}

	planningCategories := []string{}
	if timeEntry.CategoryId != nil {
		planningCategories = append(planningCategories, *timeEntry.CategoryId)
	}

	dto := &TimeEntryDTO{
		UserPlanningEventUuid: timeEntry.ActivityId,
		PlanningCategories:    planningCategories,
		StartedAt:             timeEntry.Since,
		StartTime:             timeEntry.Since,
		EndedAt:               timeEntry.Until,
		EndTime:               timeEntry.Until,
		Note:                  timeEntry.note,
		// always setting to false, not realy does anything
		IsAutomaticallyApprove: false,
	}

	payload, err := json.Marshal(*dto)
	if err != nil {
		client.logger.Error().Err(err).Msg("Error while marshaling payload")
		return err
	}

	endpointUrl := fmt.Sprintf("%s/v2/module-planning/scheduled-events", client.apiUrl)
	client.logger.Debug().Any("payload", payload).Any("DTO", dto).Str("endpoint_url", endpointUrl).Msg("Sending payload")
	req, err := http.NewRequest(http.MethodPost, endpointUrl, bytes.NewBuffer(payload))
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se udelat request")
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	client.authenticateRequest(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Error while sending request.")
		return err
	}

	if res.StatusCode != 200 {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			client.logger.Fatal().Err(err).Msg("Error while reading response body.")
			return err
		}

		message := "Non-200 response received"
		client.logger.Error().Int("status_code", res.StatusCode).Str("body", fmt.Sprintf("%s", body)).Msg(message)
		return errors.New(message)
	}

	client.logger.Info().Msg("Time entry saved")
	return nil
}
