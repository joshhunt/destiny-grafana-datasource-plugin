package plugin

import "joshhunt-destiny-datasource/pkg/query"

func validateQuery(query query.QueryModel) bool {
	if query.Profile.MembershipType == 0 {
		return false
	}

	if query.Profile.MembershipId == "" {
		return false
	}

	if query.Characters == nil {
		query.Characters = []string{}
	}

	return true
}
