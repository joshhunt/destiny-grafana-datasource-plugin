package bungieAPI

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	backend "github.com/grafana/grafana-plugin-sdk-go/backend"
	bungie "github.com/joshhunt/bungieapigo/pkg/models"
	"github.com/unknwon/log"
)

var ACTIVITIES_PAGE_SIZE = 250

type BungieAPI struct {
	apiKey string
}

func Create(apiKey string) BungieAPI {
	newInstance := BungieAPI{
		apiKey: apiKey,
	}

	return newInstance
}

func (bungieAPI BungieAPI) Get(path string, query url.Values) ([]byte, error) {
	requestUrl := path
	if !strings.Contains(requestUrl, "https://") {
		requestUrl = fmt.Sprintf("https://www.bungie.net%v", requestUrl)
	}

	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		return nil, err
	}

	if query == nil {
		query = url.Values{}
	}

	query.Set("_c", strconv.Itoa(int(time.Now().Unix())))
	req.URL.RawQuery = query.Encode()

	backend.Logger.Debug("Requesting URL", "url", req.URL.String())

	req.Header.Set("x-api-key", bungieAPI.apiKey)

	httpClient := http.Client{
		Timeout: time.Second * 10,
	}
	res, getErr := httpClient.Do(req)
	if getErr != nil {
		return nil, getErr
	}
	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return nil, readErr
	}

	return body, nil
}

func (bungieAPI BungieAPI) RequestCharacterActivityHistory(membershipType int, membershipID string, characterID string, modeType int, page int) ([]bungie.DestinyHistoricalStatsPeriodGroup, error) {
	query := url.Values{}
	query.Add("page", strconv.Itoa(page))
	query.Add("count", strconv.Itoa(ACTIVITIES_PAGE_SIZE))
	if modeType != 0 {
		query.Add("mode", strconv.Itoa(modeType))
	}

	path := fmt.Sprintf("/Platform/Destiny2/%v/Account/%v/Character/%v/Stats/Activities/", membershipType, membershipID, characterID)
	body, err := bungieAPI.Get(path, query)
	if err != nil {
		return nil, err
	}

	activityHistory := DestinyResponse[bungie.DestinyActivityHistoryResults]{}
	jsonErr := json.Unmarshal(body, &activityHistory)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return activityHistory.Response.Activities, nil
}

func (bungieAPI BungieAPI) RequestCharacterActivityHistoryForRange(membershipType int, membershipID string, characterID string, modeType int, timeRange backend.TimeRange) ([]bungie.DestinyHistoricalStatsPeriodGroup, error) {
	activities := []bungie.DestinyHistoricalStatsPeriodGroup{}
	running := true
	page := 0

	for running {
		activitiesPage, err := bungieAPI.RequestCharacterActivityHistory(membershipType, membershipID, characterID, modeType, page)
		if err != nil {
			return nil, err
		}

		page += 1

		for _, activity := range activitiesPage {
			activityStart := activity.Period
			if activityStart.Before(timeRange.From) {
				running = false
				break
			}

			if activityStart.After(timeRange.To) {
				continue
			}

			activities = append(activities, activity)
		}
	}

	return activities, nil
}

func (bungieAPI BungieAPI) RequestManifest() (*bungie.DestinyManifest, error) {
	body, err := bungieAPI.Get("/Platform/Destiny2/Manifest/", nil)
	if err != nil {
		return nil, err
	}

	data := DestinyResponse[bungie.DestinyManifest]{}
	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &data.Response, nil
}

func (bungieAPI BungieAPI) RequestSettings() (*bungie.Destiny2CoreSettings, error) {
	body, err := bungieAPI.Get("/Platform/Settings/", nil)
	if err != nil {
		return nil, err
	}

	data := DestinyResponse[bungie.Destiny2CoreSettings]{}
	jsonErr := json.Unmarshal(body, &data)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &data.Response, nil
}

func (bungieAPI BungieAPI) RequestDefinitionTable(tableName string) ([]byte, error) {
	manifest, err := bungieAPI.RequestManifest()
	if err != nil {
		log.Error("Unable to get manifest")
		return nil, err
	}

	definitionUrl := manifest.JsonWorldComponentContentPaths["en"][tableName]
	return bungieAPI.Get(definitionUrl, nil)
}

func (bungieAPI BungieAPI) RequestProfileRaw(membershipType int, membershipID string, components []int) (*DestinyResponse[bungie.DestinyProfileResponse], error) {
	query := url.Values{}
	for _, component := range components {
		query.Add("components", strconv.Itoa(component))
	}

	path := fmt.Sprintf("/Platform/Destiny2/%v/Profile/%v/", membershipType, membershipID)
	body, err := bungieAPI.Get(path, query)
	if err != nil {
		return nil, err
	}

	resp := DestinyResponse[bungie.DestinyProfileResponse]{}
	jsonErr := json.Unmarshal(body, &resp)
	if jsonErr != nil {
		return nil, jsonErr
	}

	if resp.ErrorStatus != "Success" {
		errorString := resp.ErrorStatus + ": " + resp.Message
		return nil, errors.New(errorString)
	}

	return &resp, nil
}

func (bungieAPI BungieAPI) RequestProfile(membershipType int, membershipID string, components []int) (*bungie.DestinyProfileResponse, error) {
	resp, err := bungieAPI.RequestProfileRaw(membershipType, membershipID, components)

	if err != nil {
		return nil, err
	}

	return &resp.Response, nil
}

func (bungieAPI BungieAPI) GetClassTypeName(classType bungie.DestinyClass) string {
	switch classType {
	case bungie.DestinyClassHunter:
		return "Hunter"
	case bungie.DestinyClassWarlock:
		return "Warlock"
	case bungie.DestinyClassTitan:
		return "Titan"
	case bungie.DestinyClassUnknown:
		fallthrough
	default:
		return "Unknown"
	}
}

func (bungieAPI BungieAPI) RequestCharacterDescriptions(membershipType int, membershipID string) ([]ListCharactersResourceResponseItem, error) {
	components := []int{bungie.DestinyComponentTypeCharacters}
	profile, err := bungieAPI.RequestProfile(membershipType, membershipID, components)
	if err != nil {
		backend.Logger.Error("Error requesting profile", "error", err, "membershipId", membershipID)
		return nil, err
	}

	characters := []ListCharactersResourceResponseItem{}

	if profile.Characters.Data == nil {
		backend.Logger.Warn("Characters data in profile response is empty", "profile", profile)
	} else {
		for characterId, characterData := range profile.Characters.Data {
			characterName := bungieAPI.GetClassTypeName(characterData.ClassType)
			characters = append(characters, ListCharactersResourceResponseItem{
				CharacterId: strconv.FormatInt(characterId, 10),
				Description: characterName,
			})
		}
	}

	return characters, nil
}
