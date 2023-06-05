package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	bungieAPI "destinydefinitions-destiny-datasource/pkg/bungieApi"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	bungie "github.com/joshhunt/bungieapigo/pkg/models"
	"golang.org/x/exp/slices"
)

var logger = backend.Logger

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces- only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ backend.CallResourceHandler   = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	apiKey := settings.DecryptedSecureJSONData["apiKey"]

	if apiKey == "" {
		return &Datasource{}, nil
	}

	bungieApiClient := bungieAPI.Create(apiKey)

	return &Datasource{
		bungieAPIClient: &bungieApiClient,
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	bungieAPIClient *bungieAPI.BungieAPI
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	d.bungieAPIClient = nil
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

func (d *Datasource) query(_ context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	// Unmarshal the JSON into our queryModel.
	var queryModel QueryModel

	err := json.Unmarshal(query.JSON, &queryModel)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	queryIsValid := validateQuery(queryModel)
	if !queryIsValid {
		return backend.ErrDataResponse(backend.StatusBadRequest, "Query is invalid")
	}

	allActivityHistory := []bungie.DestinyHistoricalStatsPeriodGroup{}
	includeCharacterColumn := len(queryModel.Characters) > 1
	characterDescriptions := []bungieAPI.ListCharactersResourceResponseItem{}

	if includeCharacterColumn {
		characterDescriptions, err = d.bungieAPIClient.RequestCharacterDescriptions(queryModel.Profile.MembershipType, queryModel.Profile.MembershipId)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unable to get characters: %v", err.Error()))
		}

		includeCharacterColumn = len(characterDescriptions) > 1
	}

	for _, characterId := range queryModel.Characters {
		activityHistory, err := d.bungieAPIClient.RequestCharacterActivityHistoryForRange(queryModel.Profile.MembershipType, queryModel.Profile.MembershipId, characterId, queryModel.ActivityMode, query.TimeRange)
		if err != nil {
			return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("unable to get activity history: %v", err.Error()))
		}

		characterDescriptionIndex := slices.IndexFunc(characterDescriptions, func(v bungieAPI.ListCharactersResourceResponseItem) bool { return v.CharacterId == characterId })
		var characterDescription string
		if characterDescriptionIndex > -1 {
			characterDescription = characterDescriptions[characterDescriptionIndex].Description
		}

		if includeCharacterColumn {
			for _, activity := range activityHistory {
				activity.Values["$character"] = bungie.DestinyHistoricalStatsValue{
					Basic: bungie.DestinyHistoricalStatsValuePair{
						DisplayValue: characterDescription,
					},
				}
			}
		}

		allActivityHistory = append(allActivityHistory, activityHistory...)
	}

	sort.Slice(allActivityHistory, func(i, j int) bool {
		return allActivityHistory[i].Period.After(allActivityHistory[j].Period)
	})

	timeField := data.NewField("Time", nil, []time.Time{})
	instanceIDField := data.NewField("PGCR ID", nil, []int64{})

	endTimeField := data.NewField("End time", nil, []time.Time{})
	durationField := data.NewField("Activity duration", nil, []int64{})
	timePlayedField := data.NewField("Time played", nil, []string{})

	activityModeNameField := data.NewField("Activity mode", nil, []string{})
	activityNameField := data.NewField("Activity", nil, []string{})
	directorActivityNameField := data.NewField("Director activity", nil, []string{})

	standingField := data.NewField("Standing", nil, []string{})
	completedField := data.NewField("Completed", nil, []string{})
	completionReasonField := data.NewField("Completion reason", nil, []string{})

	characterField := data.NewField("Character", nil, []string{})

	includeStanding := false

	for _, activity := range allActivityHistory {
		durationSeconds := activity.Values["activityDurationSeconds"].Basic.Value
		activityEnd := activity.Period.Add(time.Second * time.Duration(durationSeconds))
		endTimeField.Append(activityEnd)
		durationField.Append(int64(durationSeconds))

		timeField.Append(activity.Period)
		instanceIDField.Append(activity.ActivityDetails.InstanceId)

		activityModeDef := d.bungieAPIClient.GetActivityModeDefinitionForModeType(int(activity.ActivityDetails.Mode))
		activityModeNameField.Append(activityModeDef.DisplayProperties.Name)

		activityDef := d.bungieAPIClient.GetActivityDefinitionForHash(activity.ActivityDetails.ReferenceId)
		activityNameField.Append(activityDef.DisplayProperties.Name)

		directorActivityDef := d.bungieAPIClient.GetActivityDefinitionForHash(activity.ActivityDetails.DirectorActivityHash)
		directorActivityNameField.Append(directorActivityDef.DisplayProperties.Name)

		standing := activity.Values["standing"].Basic.DisplayValue
		standingField.Append(standing)

		timePlayed := activity.Values["timePlayedSeconds"].Basic.DisplayValue
		timePlayedField.Append(timePlayed)

		if standing != "" {
			includeStanding = true
		}

		completed := activity.Values["completed"].Basic.DisplayValue
		completedField.Append(completed)

		completionReason := activity.Values["completionReason"].Basic.DisplayValue
		completionReasonField.Append(completionReason)

		if includeCharacterColumn {
			characterField.Append(activity.Values["$character"].Basic.DisplayValue)
		}
	}

	// https://grafana.com/docs/grafana/latest/developers/plugins/data-frames/
	frame := data.NewFrame("response")
	frame.Fields = append(frame.Fields,
		timeField,
		endTimeField,
		durationField,
		timePlayedField,
		instanceIDField,
		activityModeNameField,
		activityNameField,
		directorActivityNameField,
		completedField,
	)

	if includeStanding {
		frame.Fields = append(frame.Fields, standingField)
	}

	if includeCharacterColumn {
		frame.Fields = append(frame.Fields, characterField)
	}

	var response backend.DataResponse
	response.Frames = append(response.Frames, frame)

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if d.bungieAPIClient == nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "API key not configured",
		}, nil
	}

	resp, err := d.bungieAPIClient.RequestProfileRaw(2, "4611686018469271298", []int{bungie.DestinyComponentTypeCharacters})

	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Health check request failed: %v", err),
		}, nil
	}

	if resp.ErrorStatus != "Success" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("%v: %v", resp.ErrorStatus, resp.Message),
		}, nil
	}

	var status = backend.HealthStatusOk
	var message = "Data source is working"

	return &backend.CheckHealthResult{
		Status:  status,
		Message: message,
	}, nil
}

func (d *Datasource) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	var resp *backend.CallResourceResponse
	var err error

	switch req.Path {
	case "profile-search":
		resp, err = d.profileSearchResourceHandler(req)
	case "list-characters":
		resp, err = d.listCharactersResourceHandler(req)
	case "list-activity-modes":
		resp, err = d.listActivityModesResourceHandler(req)
	default:
		resp = &backend.CallResourceResponse{
			Body:   []byte(`{ "message": "resource not found" }`),
			Status: http.StatusNotFound,
		}
	}

	if err != nil {
		return sender.Send(&backend.CallResourceResponse{
			Status: http.StatusInternalServerError,
			Body:   []byte(`{ "message": "unexpected error" }`),
		})
	}

	return sender.Send(resp)
}
