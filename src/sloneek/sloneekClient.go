package sloneek

import (
	"encoding/json"
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
		client.logger.Fatal().Err(err).Msg("Nepovedlo se precist categories body")
		return nil
	}

	client.logger.Info().Str("body", fmt.Sprintf("%s", body)).Msg("Prisla mi categories odpoved")
	var categoriesPayload CategoriesResponse
	err = json.Unmarshal(body, &categoriesPayload)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Unmarshal se nepovedl")
		return nil
	}

	fmt.Printf("payload.Data: %v\n", categoriesPayload.Data)
	return categoriesPayload.Data
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
	endpointUrl := fmt.Sprintf("%s/v2/module-planning/scheduled-events/options/user-planning-events", client.apiUrl)
	req, err := http.NewRequest(http.MethodGet, endpointUrl, nil)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se udelat request")
		return nil
	}

	client.authenticateRequest(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se poslat auth request")
		return nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Nepovedlo se precist body")
		return nil
	}

	client.logger.Info().Str("body", fmt.Sprintf("%s", body)).Msg("Prisla mi odpoved")
	var payload OptionsResponse
	err = json.Unmarshal(body, &payload)
	if err != nil {
		client.logger.Fatal().Err(err).Msg("Unmarshal se nepovedl")
	}

	fmt.Printf("payload.Data: %v\n", payload.Data)

	activities := make([]Activity, len(payload.Data))
	for i, item := range payload.Data {
		activities[i] = Activity{Id: item.Planning_Event.Uuid, Name: item.Planning_Event.Name}
	}

	return activities

}

func CreateSloneekClient(apiUrl string, bearerToken string, logger *zerolog.Logger) *SloneekClient {
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
