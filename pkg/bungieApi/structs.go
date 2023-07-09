package bungieAPI

import bungie "github.com/joshhunt/bungieapigo/pkg/models"

type MembershipPair struct {
	MembershipType int    `json:"membershipType"`
	MembershipId   string `json:"membershipId"`
}

type DestinyResponse[T any] struct {
	Response    T      `json:"Response"`
	ErrorCode   int    `json:"ErrorCode"`
	ErrorStatus string `json:"ErrorStatus"`
	Message     string `json:"Message"`
}

type ListCharactersResourceResponseItem struct {
	CharacterId string `json:"characterId"`
	Description string `json:"description"`
}

type ListActivityModeResourceResponseItem struct {
	Value int    `json:"value"`
	Label string `json:"label"`
}

type DestinyActivityDefinitionMap map[int]*bungie.DestinyActivityDefinition
type DestinyActivityModeDefinitionMap map[int]*bungie.DestinyActivityModeDefinition
