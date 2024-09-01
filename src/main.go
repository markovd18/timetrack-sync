package main

import (
	"encoding/json"
	//"flag"
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

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	output := os.Stderr
	logger := zerolog.New(output).With().Timestamp().Logger().Output(zerolog.ConsoleWriter{Out: output, TimeFormat: time.StampMilli})

	err := godotenv.Load(".env")
	if err != nil {
		logger.Fatal().Err(err).Msg("Error while loading environment variables")
	}

	//bearer_token := flag.String("bearer", "", "Bearer token obtained after login to Sloneek app")
	//toggl_email := flag.String("toggl-email", "", "Toggl Track login email")
	//toggl_password := flag.String("toggl-password", "", "Toggl Track password")

	//flag.Parse()

	//if *bearer_token == "" {
	//	logger.Fatal().Msg("Nezadal jsi JWT")
	//}

	//if *toggl_email == "" || *toggl_password == "" {
	//	// TODO fatal nezadal jsi kredence
	//}

	//test_api(&logger)

	//vic_test_sloneek(&logger, bearer_token)

	//auth_url := fmt.Sprintf("%s/me", TOGGL_API_URL)
	//req, err := http.NewRequest(http.MethodGet, auth_url, nil)
	//if err != nil {
	//	logger.Fatal().Err(err).Msg("Nepovedlo se vytvorit request")
	//}

	//req.SetBasicAuth(TOGGL_API_KEY, "api_token")
	//res, err := http.DefaultClient.Do(req)
	//if err != nil {
	//	logger.Fatal().Err(err).Msg("Nepovedlo se poslat auth request")
	//}
	//res.Header.Get("Set-Cookie")
	//logger.Info().Str("status", res.Status).Msg("Prisla mi repsonse")
	//if res.StatusCode == 403 {
	//	logger.Fatal().Str("reason", fmt.Sprintf("%s", read_body(res.Body, &logger))).Msg("Autentizace selhala")
	//}

	//logger.Info().Str("cookie_header", res.Header.Get("Set-Cookie")).Str("body", fmt.Sprintf("%s", read_body(res.Body, &logger))).Msg("Parsuju response")

	time_entries_url := fmt.Sprintf("%s/me/time_entries?start_date=%s&end_date=%s", TOGGL_API_URL, "2024-08-01", "2024-08-04")
	req, err := http.NewRequest(http.MethodGet, time_entries_url, nil)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se vytvorit request")
	}
	toggl_api_key := os.Getenv("TOGGL_API_KEY")
	req.SetBasicAuth(toggl_api_key, "api_token")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Fatal().Err(err).Msg("Nepovedlo se poslat auth request")
	}
	logger.Info().Str("status", res.Status).Msg("Prisla mi repsonse")
	if res.StatusCode == 403 {
		logger.Fatal().Str("reason", fmt.Sprintf("%s", read_body(res.Body, &logger))).Msg("Autentizace selhala")
	}

	logger.Info().Str("body", fmt.Sprintf("%s", read_body(res.Body, &logger))).Msg("Parsuju response")
}
