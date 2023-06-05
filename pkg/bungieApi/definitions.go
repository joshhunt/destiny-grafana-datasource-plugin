package bungieAPI

import (
	"encoding/json"

	bungie "github.com/joshhunt/bungieapigo/pkg/models"
)

var (
	cachedActivityModeDef = DestinyActivityModeDefinitionMap{}
	cachedActivityDefs    = DestinyActivityDefinitionMap{}
)

func (bungieAPI BungieAPI) initializeCachedActivityModeDef() error {
	body, err := bungieAPI.RequestDefinitionTable("DestinyActivityModeDefinition")
	if err != nil {
		return err
	}

	jsonErr := json.Unmarshal(body, &cachedActivityModeDef)
	if jsonErr != nil {
		return err
	}

	return nil
}

func (bungieAPI BungieAPI) GetActivityModeDefinitionForModeType(modeType int) *bungie.DestinyActivityModeDefinition {
	if len(cachedActivityModeDef) == 0 {
		/*err := */ bungieAPI.initializeCachedActivityModeDef()
		// if err != nil {
		// 	logger.Warn("Unable to fetch DestinyActivityModeDefinitions")
		// }
	}

	for _, def := range cachedActivityModeDef {
		if def.ModeType == bungie.DestinyActivityModeType(modeType) {
			return def
		}
	}

	// logger.Warn("Unable to find activity mode definition", "hash", modeType)
	return nil
}

func (bungieAPI BungieAPI) GetAllActivityModeDefinitions() DestinyActivityModeDefinitionMap {
	if len(cachedActivityModeDef) == 0 {
		/*err := */ bungieAPI.initializeCachedActivityModeDef()
		// if err != nil {
		//     logger.Warn("Unable to fetch DestinyActivityModeDefinitions")
		// }
	}

	return cachedActivityModeDef
}

func (bungieAPI BungieAPI) initializeCachedActivityDefs() error {
	body, err := bungieAPI.RequestDefinitionTable("DestinyActivityDefinition")
	if err != nil {
		return err
	}

	jsonErr := json.Unmarshal(body, &cachedActivityDefs)
	if jsonErr != nil {
		return err
	}

	return nil
}

func (bungieAPI BungieAPI) GetActivityDefinitionForHash(hash int) *bungie.DestinyActivityDefinition {
	if len(cachedActivityDefs) == 0 {
		/*err := */ bungieAPI.initializeCachedActivityDefs()
		// if err != nil {
		//     logger.Warn("Unable to fetch DestinyActivityDefinitions")
		// }
	}

	def := cachedActivityDefs[hash]

	// if def.Hash == 0 {
	// 	logger.Warn("Unable to find activity definition", "hash", hash)
	// }

	return def
}
