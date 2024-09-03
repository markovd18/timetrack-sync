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
	Uuid string `json:"uuid"`
	Name string `jons:"name"`
}

type CategoriesResponse struct {
	Message     string     `json:"message"`
	Status_code int32      `json:"status_code"`
	Data        []Category `json:"data"`
}

func (client *SloneekClient) GetCategories() *map[string]string {
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
	categoriesMap := make(map[string]string)
	for _, item := range categoriesPayload.Data {
		categoriesMap[item.Name] = item.Uuid
	}
	fmt.Printf("categories map: %v\n", categoriesMap)
	return &categoriesMap
}

type PlanningEvent struct {
	Uuid string `json:"uuid"`
	Name string `json:"name"`
	// pak este neco
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

func (client *SloneekClient) GetActivities() *map[string]string {
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

	activitiesMap := make(map[string]string)
	for _, item := range payload.Data {
		activitiesMap[item.Planning_Event.Name] = item.Planning_Event.Uuid
	}

	fmt.Printf("activities_map: %v\n", activitiesMap)
	return &activitiesMap

}

func CreateSloneekClient(apiUrl string, bearerToken string, logger *zerolog.Logger) *SloneekClient {
	httpClient := http.Client{Timeout: time.Minute}
	return &SloneekClient{apiUrl: apiUrl, bearerToken: bearerToken, logger: logger, httpClient: &httpClient}
}

func (client *SloneekClient) authenticateRequest(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", client.bearerToken))
}
