package validation

import (
	"net/url"
)

type LeagueData struct {
	Name      string
	ShortName string
	Sorter    int64
	Remarks   string
}

func ValidateLeague(form url.Values) (LeagueData, map[string]string) {
	errors := make(map[string]string)

	name := requireField(form, "name", "Name", errors)
	shortName := optionalField(form, "short_name")
	sorter := parseIntField(form, "sorter", "Sortierung", errors, 0)
	remarks := optionalField(form, "remarks")

	data := LeagueData{
		Name:      name,
		ShortName: shortName,
		Sorter:    sorter,
		Remarks:   remarks,
	}

	return data, errors
}
