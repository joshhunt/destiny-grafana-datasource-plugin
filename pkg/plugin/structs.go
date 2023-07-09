package plugin

import bungieAPI "joshhunt-destiny-datasource/pkg/bungieApi"

type ProfileSearchResourceRequestBody struct {
	Query string `json:"query"`
}

type ListCharactersResourceRequestBody struct {
	bungieAPI.MembershipPair
}
