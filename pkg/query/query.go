package query

import (
	"fmt"
	bungieAPI "joshhunt-destiny-datasource/pkg/bungieApi"
	"sort"
	"time"

	bungie "github.com/joshhunt/bungieapigo/pkg/models"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"golang.org/x/exp/slices"
)

type QueryModel struct {
	Characters   []string                 `json:"characters"`
	Profile      bungieAPI.MembershipPair `json:"profile"`
	ActivityMode int                      `json:"activityMode"`
}

func QueryActivityHistory(bungieAPIClient *bungieAPI.BungieAPI, dataQuery backend.DataQuery, queryModel QueryModel) (*data.Frame, error) {
	allActivityHistory := []bungie.DestinyHistoricalStatsPeriodGroup{}
	includeCharacterColumn := len(queryModel.Characters) > 1
	characterDescriptions := []bungieAPI.ListCharactersResourceResponseItem{}

	var err error

	if includeCharacterColumn {
		characterDescriptions, err = bungieAPIClient.RequestCharacterDescriptions(queryModel.Profile.MembershipType, queryModel.Profile.MembershipId)
		if err != nil {
			return nil, fmt.Errorf("unable to get characters: %v", err.Error())
		}

		includeCharacterColumn = len(characterDescriptions) > 1
	}

	for _, characterId := range queryModel.Characters {
		activityHistory, err := bungieAPIClient.RequestCharacterActivityHistoryForRange(queryModel.Profile.MembershipType, queryModel.Profile.MembershipId, characterId, queryModel.ActivityMode, dataQuery.TimeRange)
		if err != nil {
			return nil, fmt.Errorf("unable to get activity history: %v", err.Error())
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

		activityModeDef := bungieAPIClient.GetActivityModeDefinitionForModeType(int(activity.ActivityDetails.Mode))
		activityModeNameField.Append(activityModeDef.DisplayProperties.Name)

		activityDef := bungieAPIClient.GetActivityDefinitionForHash(activity.ActivityDetails.ReferenceId)
		activityNameField.Append(activityDef.DisplayProperties.Name)

		directorActivityDef := bungieAPIClient.GetActivityDefinitionForHash(activity.ActivityDetails.DirectorActivityHash)
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

	return frame, nil
}
