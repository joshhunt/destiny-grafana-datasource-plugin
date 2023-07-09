package plugin

import bungieAPI "joshhunt-destiny-datasource/pkg/bungieApi"

// TODO: move all these destiny structs elsewhere

type ProfileSearchResourceRequestBody struct {
	Query string `json:"query"`
}

type ListCharactersResourceRequestBody struct {
	bungieAPI.MembershipPair
}

// type ListCharactersResourceResponseItem struct {
// 	CharacterId string `json:"characterId"`
// 	Description string `json:"description"`
// }

// type ListActivityModeResourceResponseItem struct {
// 	Value int    `json:"value"`
// 	Label string `json:"label"`
// }
