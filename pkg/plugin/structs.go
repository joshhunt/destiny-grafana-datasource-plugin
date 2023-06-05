package plugin

type MembershipPair struct {
	MembershipType int    `json:"membershipType"`
	MembershipId   string `json:"membershipId"`
}

type QueryModel struct {
	Characters   []string       `json:"characters"`
	Profile      MembershipPair `json:"profile"`
	ActivityMode int            `json:"activityMode"`
}

// TODO: move all these destiny structs elsewhere

type ProfileSearchResourceRequestBody struct {
	Query string `json:"query"`
}

type ListCharactersResourceRequestBody struct {
	MembershipPair
}

// type ListCharactersResourceResponseItem struct {
// 	CharacterId string `json:"characterId"`
// 	Description string `json:"description"`
// }

// type ListActivityModeResourceResponseItem struct {
// 	Value int    `json:"value"`
// 	Label string `json:"label"`
// }
