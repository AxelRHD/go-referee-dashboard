package validation

import (
	"net/url"
)

var Bundeslaender = []string{
	"Baden-Württemberg",
	"Bayern",
	"Berlin",
	"Brandenburg",
	"Bremen",
	"Hamburg",
	"Hessen",
	"Mecklenburg-Vorpommern",
	"Niedersachsen",
	"Nordrhein-Westfalen",
	"Rheinland-Pfalz",
	"Saarland",
	"Sachsen",
	"Sachsen-Anhalt",
	"Schleswig-Holstein",
	"Thüringen",
}

type TeamData struct {
	Name     string
	State    string
	IsActive int64
	Remarks  string
}

func ValidateTeam(form url.Values) (TeamData, map[string]string) {
	errors := make(map[string]string)

	name := requireField(form, "name", "Name", errors)
	state := optionalField(form, "state")

	var isActive int64
	if form.Get("is_active") != "" {
		isActive = 1
	}

	remarks := optionalField(form, "remarks")

	data := TeamData{
		Name:     name,
		State:    state,
		IsActive: isActive,
		Remarks:  remarks,
	}

	return data, errors
}
