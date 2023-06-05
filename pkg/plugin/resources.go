package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	bungieAPI "destinydefinitions-destiny-datasource/pkg/bungieApi"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func (d *Datasource) profileSearchResourceHandler(req *backend.CallResourceRequest) (*backend.CallResourceResponse, error) {
	requestBody := ProfileSearchResourceRequestBody{}
	err := json.Unmarshal(req.Body, &requestBody)
	if err != nil {
		logger.Error("Unable to unmarshal profile search body", "error", err)
		return nil, err
	}

	searchUrl := fmt.Sprintf("https://elastic.destinytrialsreport.com/players/0/%v/", url.PathEscape(requestBody.Query))
	data, err := d.bungieAPIClient.Get(searchUrl, nil)
	if err != nil {
		logger.Error("Unable call to DTR search failed", "error", err)
		return nil, err
	}

	resp := &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   data,
	}

	return resp, nil
}

func (d *Datasource) listCharactersResourceHandler(req *backend.CallResourceRequest) (*backend.CallResourceResponse, error) {
	requestBody := ListCharactersResourceRequestBody{}
	err := json.Unmarshal(req.Body, &requestBody)
	if err != nil {
		logger.Error("Unable to unmarshal listCharacters body", "error", err)
		return nil, err
	}

	characters, err := d.bungieAPIClient.RequestCharacterDescriptions(requestBody.MembershipType, requestBody.MembershipId)
	if err != nil {
		logger.Error("Error requesting character descriptions", "error", err, "membershipId", requestBody.MembershipId)
		return nil, err
	}

	respBody, err := json.Marshal(characters)
	if err != nil {
		logger.Error("Unable to marshal listCharactersResourceHandler response", "error", err)
		return nil, err
	}

	resp := &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   respBody,
	}

	return resp, nil
}

func (d *Datasource) listActivityModesResourceHandler(req *backend.CallResourceRequest) (*backend.CallResourceResponse, error) {
	allDefs := d.bungieAPIClient.GetAllActivityModeDefinitions()

	activityModes := make([]bungieAPI.ListActivityModeResourceResponseItem, 0, len(allDefs))

	for _, activityMode := range allDefs {
		item := bungieAPI.ListActivityModeResourceResponseItem{
			Value: int(activityMode.ModeType),
			Label: activityMode.DisplayProperties.Name,
		}
		activityModes = append(activityModes, item)
	}

	respBody, err := json.Marshal(activityModes)
	if err != nil {
		logger.Error("Unable to marshal listActivityModesResourceHandler response", "error", err)
		return nil, err
	}

	resp := &backend.CallResourceResponse{
		Status: http.StatusOK,
		Body:   respBody,
	}

	return resp, nil
}
